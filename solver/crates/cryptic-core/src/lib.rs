use serde::Serialize;
use std::collections::HashMap;
use std::fmt;

#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
#[serde(rename_all = "snake_case")]
pub enum Mechanism {
    Anagram,
    Pattern,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct Candidate {
    pub answer: String,
    pub mechanism: Mechanism,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub fodder: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub pattern: Option<String>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Error {
    EmptyLetters,
    EmptyPattern,
    InvalidPatternCharacter(char),
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::EmptyLetters => write!(f, "letters must contain at least one letter or digit"),
            Self::EmptyPattern => write!(f, "pattern must not be empty"),
            Self::InvalidPatternCharacter(c) => {
                write!(f, "pattern contains unsupported character {c:?}")
            }
        }
    }
}

impl std::error::Error for Error {}

#[derive(Debug, Clone)]
pub struct Wordlist {
    words_by_length: HashMap<usize, Vec<String>>,
    words_by_signature: HashMap<String, Vec<String>>,
}

impl Wordlist {
    pub fn parse(input: &str) -> Self {
        let mut words: Vec<String> = input.lines().filter_map(normalize_word).collect();
        words.sort();
        words.dedup();

        let mut words_by_length: HashMap<usize, Vec<String>> = HashMap::new();
        let mut words_by_signature: HashMap<String, Vec<String>> = HashMap::new();

        for word in words {
            words_by_length
                .entry(word.chars().count())
                .or_default()
                .push(word.clone());
            words_by_signature
                .entry(signature(&word))
                .or_default()
                .push(word);
        }

        Self {
            words_by_length,
            words_by_signature,
        }
    }

    pub fn anagrams(&self, letters: &str, enumeration: usize) -> Result<Vec<Candidate>, Error> {
        let normalized = normalize_letters(letters);
        if normalized.is_empty() {
            return Err(Error::EmptyLetters);
        }
        if enumeration > 0 && normalized.chars().count() != enumeration {
            return Ok(Vec::new());
        }

        let candidates = self
            .words_by_signature
            .get(&signature(&normalized))
            .into_iter()
            .flatten()
            .map(|word| Candidate {
                answer: word.clone(),
                mechanism: Mechanism::Anagram,
                fodder: Some(normalized.clone()),
                pattern: None,
            })
            .collect();
        Ok(candidates)
    }

    pub fn pattern_matches(
        &self,
        pattern: &str,
        enumeration: usize,
    ) -> Result<Vec<Candidate>, Error> {
        let pattern = normalize_pattern(pattern)?;
        let length = if enumeration == 0 {
            pattern.chars().count()
        } else {
            enumeration
        };
        if pattern.chars().count() != length {
            return Ok(Vec::new());
        }

        let candidates = self
            .words_by_length
            .get(&length)
            .into_iter()
            .flatten()
            .filter(|word| pattern_matches_word(&pattern, word))
            .map(|word| Candidate {
                answer: word.clone(),
                mechanism: Mechanism::Pattern,
                fodder: None,
                pattern: Some(pattern.clone()),
            })
            .collect();
        Ok(candidates)
    }
}

fn normalize_word(input: &str) -> Option<String> {
    let word = normalize_letters(input);
    (!word.is_empty()).then_some(word)
}

fn normalize_letters(input: &str) -> String {
    input
        .chars()
        .filter(|c| c.is_alphanumeric())
        .flat_map(char::to_lowercase)
        .collect()
}

fn normalize_pattern(input: &str) -> Result<String, Error> {
    let trimmed = input.trim();
    if trimmed.is_empty() {
        return Err(Error::EmptyPattern);
    }

    trimmed
        .chars()
        .map(|c| match c {
            '?' | '.' => Ok('?'),
            c if c.is_alphanumeric() => Ok(c.to_ascii_lowercase()),
            c => Err(Error::InvalidPatternCharacter(c)),
        })
        .collect()
}

fn signature(input: &str) -> String {
    let mut chars: Vec<char> = input.chars().collect();
    chars.sort_unstable();
    chars.into_iter().collect()
}

fn pattern_matches_word(pattern: &str, word: &str) -> bool {
    pattern
        .chars()
        .zip(word.chars())
        .all(|(expected, actual)| expected == '?' || expected == actual)
}

#[cfg(test)]
mod tests {
    use super::*;

    fn words() -> Wordlist {
        Wordlist::parse("cat\ncater\ncrate\ntrace\nreact\ncream\n")
    }

    #[test]
    fn finds_exact_anagrams_in_stable_order() {
        let results = words().anagrams("CARET", 5).unwrap();
        let answers: Vec<_> = results.iter().map(|c| c.answer.as_str()).collect();
        assert_eq!(answers, ["cater", "crate", "react", "trace"]);
        assert!(results.iter().all(|c| c.mechanism == Mechanism::Anagram));
    }

    #[test]
    fn anagram_enumeration_must_match_fodder_length() {
        assert!(words().anagrams("caret", 4).unwrap().is_empty());
    }

    #[test]
    fn matches_known_crossing_letters_case_insensitively() {
        let results = words().pattern_matches("?R?C?", 5).unwrap();
        let answers: Vec<_> = results.iter().map(|c| c.answer.as_str()).collect();
        assert_eq!(answers, ["trace"]);
    }

    #[test]
    fn pattern_length_can_be_inferred() {
        let results = words().pattern_matches("c?t", 0).unwrap();
        assert_eq!(results[0].answer, "cat");
    }

    #[test]
    fn rejects_pattern_punctuation_instead_of_treating_it_as_regex() {
        assert_eq!(
            words().pattern_matches("c*t", 3),
            Err(Error::InvalidPatternCharacter('*'))
        );
    }
}
