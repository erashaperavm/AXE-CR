use anyhow::Result;
use clap::Parser;
use sp1_sdk::{ProverClient, HashableKey, Prover, ProvingKey};
use std::fs;
use std::path::PathBuf;

#[derive(Parser)]
struct Cli {
    /// ELF file of the program that generated the proof
    #[arg(short, long)]
    elf: PathBuf,

    /// Path to the proof file
    #[arg(short, long)]
    proof: PathBuf,
}

#[tokio::main]
async fn main() -> Result<()> {
    env_logger::init();
    let cli = Cli::parse();

    let elf_bytes = fs::read(&cli.elf)?;
    let proof = sp1_sdk::SP1ProofWithPublicValues::load(&cli.proof)?;

    let client = ProverClient::from_env().await;
    let pk = client.setup(elf_bytes.into()).await?;
    client.verify(&proof, pk.verifying_key(), None)?;

    let vkey_str = pk.verifying_key().bytes32();
    let pv_hex = hex::encode(proof.public_values.as_slice());
    let summary = serde_json::json!({
        "status": "success",
        "vkey": vkey_str,
        "public_values_hex": pv_hex
    });
    println!("{}", serde_json::to_string(&summary)?);

    Ok(())
}