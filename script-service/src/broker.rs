use pulsar::{Pulsar, TokioExecutor, Producer, producer::SendFuture, Error as PulsarError, Consumer, SubType};

use crate::messages::TestData;

pub struct Broker {
    node_name: String,
    pulsar: Pulsar<TokioExecutor>,
    producer: Producer<TokioExecutor>,
    pub consumer: Consumer<TestData, TokioExecutor>
}

impl Broker {
    pub async fn new(addr: String, node_name: String) -> Self {
        let pulsar = Pulsar::builder(addr, TokioExecutor).build().await.expect("Could not connect to pulsar!");

        let producer = pulsar.producer()
        .with_topic("non-persistent://public/default/test")
        .with_name(node_name.clone())
        .build().await.expect("Could not create producer");

        let consumer = pulsar
        .consumer()
        .with_topic("non-persistent://public/default/test")
        .with_consumer_name(node_name.clone())
        .with_subscription_type(SubType::Exclusive)
        .with_subscription(node_name.clone())
        .build()
        .await.expect("Could not build consumer");

        Broker {
            node_name,
            pulsar,
            producer,
            consumer
        }
    }

    pub async fn send(&mut self, data: TestData) -> Result<SendFuture, PulsarError> {
        self.producer.send(data).await
    }
}