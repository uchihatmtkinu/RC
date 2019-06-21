#include "paillier.h"

# define SECURITY_PARAM 1024 

typedef struct KEYS
{
	int modulubits; // equal to security parameter
	paillier_prvkey_t* prvkey; // private key
	paillier_pubkey_t* pubkey; // public key
}paillier_keys;

void 
key_gen(paillier_keys* keys);

paillier_ciphertext_t* 
enc_add_circuit(
	paillier_pubkey_t* pubkey, 
	paillier_plaintext_t* message1, 
	paillier_plaintext_t* message2);

paillier_plaintext_t* 
dec_add_circuit(
	paillier_keys* keys ,
	paillier_ciphertext_t* cipher3);

paillier_ciphertext_t* 
enc_mul_circuit(
	paillier_pubkey_t* pubkey, 
	paillier_plaintext_t* message1, 
	paillier_plaintext_t* message2, 
	paillier_ciphertext_t* secret);

paillier_plaintext_t* 
dec_mul_circuit(
	paillier_keys* keys,
	paillier_ciphertext_t* cipher,
	paillier_ciphertext_t* secret);
