use clap::{Parser, Subcommand};
use cryptic_core::{Candidate, Wordlist};
use serde::Serialize;
use std::{fs, path::PathBuf, process::ExitCode};

#[derive(Debug, Parser)]
#[command(name = "cryptic", about = "Candidate tools for cryptic crosswords")]
struct Cli {
    #[arg(long, global = true, default_value = "../wordlists/english.txt")]
    wordlist: PathBuf,

    #[command(subcommand)]
    command: Command,
}

#[derive(Debug, Subcommand)]
enum Command {
    /// Find exact anagrams of the supplied fodder.
    Anagram {
        #[arg(long)]
        letters: String,
        #[arg(long, default_value_t = 0)]
        enumeration: usize,
    },
    /// Find words compatible with known crossing letters (? is unknown).
    Pattern {
        #[arg(long)]
        known: String,
        #[arg(long, default_value_t = 0)]
        enumeration: usize,
    },
}

#[derive(Debug, Serialize)]
struct Output {
    candidates: Vec<Candidate>,
}

fn main() -> ExitCode {
    match run(Cli::parse()) {
        Ok(output) => {
            println!("{}", serde_json::to_string_pretty(&output).unwrap());
            ExitCode::SUCCESS
        }
        Err(error) => {
            eprintln!("cryptic: {error}");
            ExitCode::FAILURE
        }
    }
}

fn run(cli: Cli) -> Result<Output, Box<dyn std::error::Error>> {
    let contents = fs::read_to_string(&cli.wordlist).map_err(|error| {
        format!(
            "could not read word list {}: {error}",
            cli.wordlist.display()
        )
    })?;
    let wordlist = Wordlist::parse(&contents);

    let candidates = match cli.command {
        Command::Anagram {
            letters,
            enumeration,
        } => wordlist.anagrams(&letters, enumeration)?,
        Command::Pattern { known, enumeration } => wordlist.pattern_matches(&known, enumeration)?,
    };

    Ok(Output { candidates })
}
