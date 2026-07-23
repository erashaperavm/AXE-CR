use anyhow::Result;
use clap::{Parser};
use sp1_sdk::{ProverClient, SP1Stdin, SP1ProofMode, Prover, ProveRequest, ProvingKey, HashableKey};
use std::fs;
use std::path::PathBuf;

#[derive(Parser)]
struct Cli {
    /// 标准格式的 AXE RsFunc 编译产物路径（包含 ELF 文件）
    #[arg(short, long)]
    cwkdir: PathBuf,

    /// 项目名称
    #[arg(short, long)]
    fnname: String,

    /// 输入文件路径 (可选)
    #[arg(short, long)]
    input: Option<PathBuf>,

    /// 证明模式: core, plonk, groth16, compressed
    #[arg(short, long, default_value = "core")]
    mode: String,

    /// 工作目录（输出文件保存在这里）
    #[arg(short, long)]
    wkdir: PathBuf,
}

#[tokio::main]
async fn main() -> Result<()> {
    env_logger::init();
    let cli = Cli::parse();

    run_prove(&cli.cwkdir, &cli.fnname, cli.input.as_ref(), &cli.mode, &cli.wkdir).await?;

    Ok(())
}

async fn run_prove(
    target_path: &PathBuf,
    function_name: &String,
    input_path: Option<&PathBuf>,
    mode: &str,
    wkdir: &PathBuf
) -> Result<()> {
    // 1. 读取 ELF 文件
    let elf_bytes = fs::read(PathBuf::from(target_path).join("elf-compilation").join("riscv64im-succinct-zkvm-elf").join("release").join(function_name))?;

    // 2. 准备输入数据 (stdin)
    let (stdin_exec, stdin_prove) = if let Some(path) = input_path {
        let data = fs::read(path)?;
        (SP1Stdin::from(&data), SP1Stdin::from(&data))
    } else {
        (SP1Stdin::new(), SP1Stdin::new())
    };

    // 3. 初始化客户端并设置程序
    let client = ProverClient::from_env().await;
    let pk = client.setup(elf_bytes.into()).await?;

    // 执行程序以获取公开输出和执行报告
    let (public_values, exec_report) = client.execute(pk.elf().clone(), stdin_exec).await?;

    // 设置路径
    let base:PathBuf = PathBuf::from(wkdir).join("execution_out");
    let pv_out:PathBuf = base.join("pv.txt");
    let rp_out:PathBuf = base.join("report.json");
    let pf_out:PathBuf = base.join("proof.bin");
    let vkey_out:PathBuf = base.join("vkey.txt");

    // 保存公开输出（十六进制）和执行报告（JSON）到文件
    let pv_text = hex::encode(public_values.as_slice());
    fs::write(pv_out, pv_text)?;
    let rp_text = serde_json::to_string(&exec_report).unwrap_or_else(|_| format!("{:#?}", exec_report));
    fs::write(rp_out, rp_text)?;

    // 4. 根据模式生成证明
    let proof_mode = match mode.to_lowercase().as_str() {
        "core" => SP1ProofMode::Core,
        "compressed" => SP1ProofMode::Compressed,
        "plonk" => SP1ProofMode::Plonk,
        "groth16" => SP1ProofMode::Groth16,
        _ => anyhow::bail!("unsupported mode: {}", mode),
    };
    let proof = client.prove(&pk, stdin_prove).mode(proof_mode).await?;

    // 保存证明文件
    proof.save(&pf_out)?;

    // 保存验证密钥
    let vkey_str = pk.verifying_key().bytes32();
    fs::write(vkey_out, &vkey_str)?;

    // 6. 输出摘要到 stdout
    let summary = serde_json::json!({
        "status": "success",
    });
    println!("{}", serde_json::to_string(&summary)?);

    Ok(())
}