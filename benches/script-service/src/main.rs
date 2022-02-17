use std::{time::Duration, sync::{Arc, atomic::{AtomicU32, Ordering}}};
use log::info;
use tokio::{sync::{mpsc::{self, Sender, Receiver}}, time::{Instant, sleep, self}, process::Command};
use tokio_tungstenite::{connect_async, tungstenite::{Message, Error}};
use futures::{StreamExt, SinkExt};
use std::env;

async fn give_results(mut rx: Receiver<bool>){
    let mut err_count = Arc::new(AtomicU32::new(0));
    let mut success_count = Arc::new(AtomicU32::new(0));

    let display_err = err_count.clone();
    let display_success = success_count.clone();
    let display = tokio::spawn(async move {
        sleep(Duration::from_millis(1000)).await;

        let mut interval = time::interval(Duration::from_millis(1000));
        let mut prev_err = 0;
        let mut prev_succ = 0;
        let mut curr_err = 0;
        let mut curr_succ = 0;
        let mut prev_time = Instant::now();
        let start = Instant::now();

        loop {
            interval.tick().await;

            prev_err = curr_err;
            prev_succ = curr_succ;
            curr_err = display_err.load(Ordering::Relaxed);
            curr_succ = success_count.load(Ordering::Relaxed);
            let elapsed = prev_time.elapsed().as_millis();
            let rate: f64 = ((curr_err + curr_succ - prev_err - prev_succ) as f64 / (elapsed as f64)) as f64 * 1000 as f64;
            Command::new("clear").status().await;
            info!("Errors: {}", curr_err);
            info!("Success: {}", curr_succ);
            info!("Current rate: {:.2} / s", rate);
            info!("Elapsed time: {}s\n", start.elapsed().as_secs());
            prev_time = Instant::now();
        }
    });

    let forever = tokio::spawn(async move {
        info!("Starting result listener...");
        loop {
            let res = rx.recv().await;
            if res.unwrap() {
                err_count.fetch_add(1, Ordering::Relaxed);
            } else {
                display_success.fetch_add(1, Ordering::Relaxed);
            }
        }
    });

    forever.await;
    display.await;
}

async fn send_reqs(tx: Sender<bool>, mut rate: u32) {
    sleep(Duration::from_millis(1000)).await;

    info!("Starting benchmark...");

    let mut last_sent = Instant::now();
    let mut count = 0;
    // rate = rate / 100;
    loop {
        let sender = tx.clone();
        tokio::spawn(async move {
            let (ws_stream, _) = connect_async("ws://localhost:3000").await.expect("Failed to connect");
            let (mut ws_sender, _) = ws_stream.split();
            let mut res: Result<(), Error> = Ok(());
            for _ in 0..100 {
                res = ws_sender.send(Message::Text("Hello World!".to_string())).await;
            }
            ws_sender.send(Message::Close(None)).await;
            ws_sender.close().await;
            sender.send(res.is_err()).await;
        });

        count += 1;

        if count > rate {
            count = 0;
            let time_taken = Instant::now().duration_since(last_sent).as_millis();
            if time_taken < 1000 {
                sleep(Duration::from_millis((1000 - time_taken).try_into().unwrap())).await;
            }

            last_sent = Instant::now();
        }
    }
}

fn main() {
    env_logger::init();

    let (tx, rx) = mpsc::channel::<bool>(10000);
    
    let rt = tokio::runtime::Builder::new_multi_thread()
        .enable_all()
        .build()
        .unwrap();

    let rate_str = env::var("rate").expect("Rate not provided!");
    let rate: u32 = rate_str.parse().expect("Invalid rate!");

    rt.spawn(send_reqs(tx, rate));
    rt.block_on(give_results(rx));
}