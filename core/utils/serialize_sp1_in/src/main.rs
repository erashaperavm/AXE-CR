use sp1_sdk::SP1Stdin;
use std::fs;
use std::path::PathBuf;
use clap::Parser;
use anyhow::Result;
use serde_json::json;

#[derive(Parser)]
struct Cli {
    #[arg(long)]
    swkdir: PathBuf,

    #[arg(long)]
    int: Vec<i64>,

    #[arg(long)]
    byte: Vec<String>,
}

fn main() -> Result<()> {
    sp1_sdk::utils::setup_logger();
    let cli = Cli::parse();

    // 将十六进制字符串转回 Vec<Vec<u8>>
    let bytes: Vec<Vec<u8>> = cli.byte
        .iter()
        .map(|s| hex::decode(s).unwrap())
        .collect();

    match serialize_sp1_in(&cli.int, &bytes, &cli.swkdir) {
        Ok(path) => {
            let response = json!({
                "status": "success",
                "message": "Serialization completed",
                "path": path
            });
            println!("{}", serde_json::to_string(&response)?);
        }
        Err(e) => {
            let response = json!({
                "status": "error",
                "message": e.to_string()
            });
            println!("{}", serde_json::to_string(&response)?);
        }
    }

    Ok(())
}

fn serialize_sp1_in(ints: &Vec<i64>, bytes: &Vec<Vec<u8>>, swkdir: &PathBuf) -> Result<String> {
    let mut stdin = SP1Stdin::new();

    stdin.write(&ints);
    stdin.write(&bytes);

    let encoded: Vec<u8> = bincode::serialize(&stdin)?;
    let path: PathBuf = swkdir.join("input.bin");
    fs::write(&path, encoded)?;

    Ok(path.to_string_lossy().to_string())
}