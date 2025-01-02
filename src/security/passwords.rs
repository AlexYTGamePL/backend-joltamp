use argon2::{password_hash::{
    rand_core::OsRng,
    PasswordHash, PasswordHasher, PasswordVerifier, SaltString
}, Argon2};

// type DynError = Box<dyn argon2::password_hash::Error + Send + Sync>;
pub fn hash_password(password: &str) -> Result<String, argon2::password_hash::Error> {
    let salt = SaltString::generate(&mut OsRng);
    let argon2 = Argon2::default();
    let hash = argon2.hash_password(password.as_bytes(), &salt)?;
    Ok(hash.to_string())
}

pub fn verify_password(password: &str, hashed_password: &str) -> Result<(), String> {
    let parsed_hash = PasswordHash::new(hashed_password).map_err(|err| {
        format!("Failed to parse hashed password: {}", err)
    })?;

    Argon2::default().verify_password(password.as_bytes(), &parsed_hash).map_err(|err| {
        format!("Password verification failed: {}", err)
    })
}