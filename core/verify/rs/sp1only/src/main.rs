use anyhow::Result;
use clap::Parser;
use sp1_sdk::{ProverClient, Prover, ProvingKey};
use std::path::PathBuf;
use std::fs;

#[derive(Parser)]
struct Cli {
    /// RsFunc 编译产物目录
    #[arg(long, short)]
    cwkdir: PathBuf,

    /// 运行输出目录
    ewkdir: PathBuf,

    /// function name
    #[arg(long, short)]
    fnname: String,
}

#[tokio::main]
async fn main() -> Result<()> {
    env_logger::init();
    let cli = Cli::parse();

    run_verify(&cli.cwkdir, &cli.ewkdir, &cli.fnname).await?;

    Ok(())
}

async fn run_verify(
    compile_wkdir: &PathBuf,
    exec_wkdir: &PathBuf,
    function_name: &String,
) -> Result<()> {
    let pf_path:PathBuf = PathBuf::from(exec_wkdir).join("execution_out").join("proof.bin");

    let proof = sp1_sdk::SP1ProofWithPublicValues::load(pf_path)?;
    let elf_bytes = fs::read(PathBuf::from(compile_wkdir).join("elf-compilation").join("riscv64im-succinct-zkvm-elf").join("release").join(function_name))?;

    let client = ProverClient::from_env().await;
    let pk = client.setup(elf_bytes.into()).await?;
    client.verify(&proof, pk.verifying_key(), None)?;

    let summary = serde_json::json!({
        "status": "success",
    });
    println!("{}", serde_json::to_string(&summary)?);

    Ok(())
}