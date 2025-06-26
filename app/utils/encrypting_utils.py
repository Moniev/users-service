from app.config.settings import cached_settings, get_settings
from cryptography.hazmat.primitives.ciphers import Cipher, algorithms, modes
from cryptography.hazmat.backends import default_backend
from os import urandom


def encrypt_data(data: bytes) -> bytes:
    nonce = urandom(12) 
    cipher = Cipher(algorithms.ChaCha20(get_settings().JWT, nonce), modes.Poly1305(), backend=default_backend())
    encryptor = cipher.encryptor()
    ciphertext = encryptor.update(data) + encryptor.finalize()
    return nonce + encryptor.tag + ciphertext 


def decrypt_data(encrypted_data: bytes) -> bytes:
    if len(encrypted_data) < 28: 
        raise ValueError("Encrypted data too short to contain nonce and tag.")

    nonce = encrypted_data[:12]
    tag = encrypted_data[12:28]
    ciphertext = encrypted_data[28:]

    cipher = Cipher(algorithms.ChaCha20(get_settings(), nonce), modes.Poly1305(), backend=default_backend())
    decryptor = cipher.decryptor()
    
    plaintext = decryptor.update(ciphertext) + decryptor.finalize()
    
    if tag != decryptor.tag:
        raise ValueError("Invalid authentication tag - data has been tampered with or incorrect key/nonce.")
    
    return plaintext