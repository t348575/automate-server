pub mod utils {
    pub fn load_env() -> Result<(), String> {
        let mut found = false;
        for (i, arg) in std::env::args().enumerate() {
            if arg.trim() == "--help" {
                println!("Usage:\n\t[-dev] run in dev mode\n\t[-env] the .env file to be used");
                std::process::exit(0);
            }

            if arg.trim() == "-dev" {
                std::env::set_var("RUST_LOG", "debug");
            }

            if arg.trim() == "-env" {
                let args: Vec<String> = std::env::args().collect();

                if i + 1 >= args.len() {
                    println!("-env path not provided");
                    std::process::exit(1);
                } else {
                    if let Err(e) = dotenv::from_path(args[i + 1].clone()) {
                        return Err(e.to_string());
                    }
                    found = true;
                }
            }
        }

        if !found {
            if let Err(e) = dotenv::dotenv() {
                return Err(e.to_string());
            }
        }

        Ok(())
    }

    pub fn parse_config<T: serde::de::DeserializeOwned>()-> Result<T, String> {
        match envy::from_env::<T>() {
            Ok(config) => Ok(config),
            Err(error) => Err(error.to_string())
        }
    }
}