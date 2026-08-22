use clap::{Parser, Subcommand};
use cryptic_core::{Analysis, Candidate, Wordlist};
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
    /// Suggest structured wordplay parses for a complete clue.
    Analyse {
        #[arg(long)]
        clue: String,
        /// Known crossing letters, using ? for an unknown letter.
        #[arg(long)]
        known: Option<String>,
    },
}

#[derive(Debug, Serialize)]
#[serde(untagged)]
enum Output {
    Candidates { candidates: Vec<Candidate> },
    Analysis(Analysis),
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

    let output = match cli.command {
        Command::Anagram {
            letters,
            enumeration,
        } => Output::Candidates {
            candidates: wordlist.anagrams(&letters, enumeration)?,
        },
        Command::Pattern { known, enumeration } => Output::Candidates {
            candidates: wordlist.pattern_matches(&known, enumeration)?,
        },
        Command::Analyse { clue, known } => {
            Output::Analysis(wordlist.analyse(&clue, known.as_deref())?)
        }
    };

    Ok(output)
}
