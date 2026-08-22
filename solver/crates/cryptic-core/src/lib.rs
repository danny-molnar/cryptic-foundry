use serde::Serialize;
use std::collections::HashMap;
use std::fmt;

const ANAGRAM_INDICATORS: &[&str] = &[
    "broken", "confused", "dancing", "drunk", "mixed", "out", "strange", "wild",
];

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
    #[serde(skip_serializing_if = "Option::is_none")]
    pub indicator: Option<String>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub matches_pattern: Option<bool>,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct Enumeration {
    pub raw: String,
    pub parts: Vec<usize>,
    pub total: usize,
}

#[derive(Debug, Clone, PartialEq, Eq, Serialize)]
pub struct Analysis {
    pub clue: String,
    pub enumeration: Enumeration,
    pub candidates: Vec<Candidate>,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub enum Error {
    EmptyLetters,
    EmptyPattern,
    InvalidPatternCharacter(char),
    MissingEnumeration,
    InvalidEnumeration(String),
}

impl fmt::Display for Error {
    fn fmt(&self, f: &mut fmt::Formatter<'_>) -> fmt::Result {
        match self {
            Self::EmptyLetters => write!(f, "letters must contain at least one letter or digit"),
            Self::EmptyPattern => write!(f, "pattern must not be empty"),
            Self::InvalidPatternCharacter(c) => {
                write!(f, "pattern contains unsupported character {c:?}")
            }
            Self::MissingEnumeration => write!(f, "clue must end with an enumeration such as (5)"),
            Self::InvalidEnumeration(value) => write!(f, "invalid enumeration {value:?}"),
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
                indicator: None,
                matches_pattern: None,
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
                indicator: None,
                matches_pattern: None,
            })
            .collect();
        Ok(candidates)
    }

    pub fn analyse(&self, clue: &str, known: Option<&str>) -> Result<Analysis, Error> {
        let (clue_text, enumeration) = parse_enumeration(clue)?;
        let pattern = known.map(normalize_pattern).transpose()?;
        if pattern
            .as_ref()
            .is_some_and(|value| value.chars().count() != enumeration.total)
        {
            return Err(Error::InvalidEnumeration(format!(
                "pattern length does not match {}",
                enumeration.raw
            )));
        }

        let words: Vec<String> = clue_text
            .split_whitespace()
            .map(normalize_letters)
            .filter(|word| !word.is_empty())
            .collect();
        let mut candidates = Vec::new();

        for (indicator_index, indicator) in words.iter().enumerate() {
            if !ANAGRAM_INDICATORS.contains(&indicator.as_str()) {
                continue;
            }

            for fodder in adjacent_fodder(&words, indicator_index, enumeration.total) {
                for mut candidate in self.anagrams(&fodder, enumeration.total)? {
                    candidate.indicator = Some(indicator.clone());
                    candidate.matches_pattern = pattern
                        .as_ref()
                        .map(|value| pattern_matches_word(value, &candidate.answer));
                    candidate.pattern = pattern.clone();
                    candidates.push(candidate);
                }
            }
        }

        candidates.sort_by(|left, right| {
            right
                .matches_pattern
                .cmp(&left.matches_pattern)
                .then_with(|| left.answer.cmp(&right.answer))
                .then_with(|| left.fodder.cmp(&right.fodder))
        });
        candidates.dedup();

        Ok(Analysis {
            clue: clue_text.to_owned(),
            enumeration,
            candidates,
        })
    }
}

fn parse_enumeration(clue: &str) -> Result<(&str, Enumeration), Error> {
    let clue = clue.trim();
    let open = clue.rfind('(').ok_or(Error::MissingEnumeration)?;
    if !clue.ends_with(')') {
        return Err(Error::MissingEnumeration);
    }

    let raw = &clue[open + 1..clue.len() - 1];
    if raw.is_empty() || raw.starts_with([',', '-']) || raw.ends_with([',', '-']) {
        return Err(Error::InvalidEnumeration(raw.to_owned()));
    }

    let mut parts = Vec::new();
    for value in raw.split([',', '-']) {
        let value = value.trim();
        let part = value
            .parse::<usize>()
            .map_err(|_| Error::InvalidEnumeration(raw.to_owned()))?;
        if part == 0 {
            return Err(Error::InvalidEnumeration(raw.to_owned()));
        }
        parts.push(part);
    }
    let total = parts.iter().sum();

    Ok((
        clue[..open].trim(),
        Enumeration {
            raw: raw.to_owned(),
            parts,
            total,
        },
    ))
}

fn adjacent_fodder(words: &[String], indicator: usize, length: usize) -> Vec<String> {
    let mut results = Vec::new();

    let mut before = String::new();
    for index in (0..indicator).rev() {
        before = format!("{}{before}", words[index]);
        match before.chars().count().cmp(&length) {
            std::cmp::Ordering::Equal => {
                results.push(before);
                break;
            }
            std::cmp::Ordering::Greater => break,
            std::cmp::Ordering::Less => {}
        }
    }

    let mut after = String::new();
    for word in &words[indicator + 1..] {
        after.push_str(word);
        match after.chars().count().cmp(&length) {
            std::cmp::Ordering::Equal => {
                results.push(after);
                break;
            }
            std::cmp::Ordering::Greater => break,
            std::cmp::Ordering::Less => {}
        }
    }

    results
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

    #[test]
    fn parses_multi_part_enumerations() {
        let (_, enumeration) = parse_enumeration("Odd phrase (3, 4-5)").unwrap();
        assert_eq!(enumeration.parts, [3, 4, 5]);
        assert_eq!(enumeration.total, 12);
    }

    #[test]
    fn analyses_anagram_and_ranks_crossing_match_first() {
        let result = words()
            .analyse("Confused caret produces a response (5)", Some("r??c?"))
            .unwrap();

        assert_eq!(result.enumeration.total, 5);
        assert_eq!(result.candidates[0].answer, "react");
        assert_eq!(result.candidates[0].indicator.as_deref(), Some("confused"));
        assert_eq!(result.candidates[0].fodder.as_deref(), Some("caret"));
        assert_eq!(result.candidates[0].matches_pattern, Some(true));
    }

    #[test]
    fn finds_multi_word_fodder_adjacent_to_indicator() {
        let result = words()
            .analyse("Cat er broken for diner (5)", None)
            .unwrap();
        assert!(result.candidates.iter().any(|item| item.answer == "cater"));
    }

    #[test]
    fn analysis_requires_terminal_enumeration() {
        assert_eq!(
            words().analyse("Confused caret", None),
            Err(Error::MissingEnumeration)
        );
    }
}
