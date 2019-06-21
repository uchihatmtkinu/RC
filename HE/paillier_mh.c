#include <stdio.h>
#include <stdlib.h>
#include <gmp.h>
#include "paillier_mh.h"

/*
	Generate public key and private key
 */
void 
key_gen(paillier_keys* keys)
{
	keys -> modulubits = SECURITY_PARAM;
	paillier_keygen(keys->modulubits, &(keys->pubkey), &(keys->prvkey), paillier_get_rand_devurandom);
}


/*
	Apply additive encryption circuit
	Enc(m1) + Enc(m2) mod n;
 */
paillier_ciphertext_t* 
enc_add_circuit(
	paillier_pubkey_t* pubkey, 
	paillier_plaintext_t* message1, 
	paillier_plaintext_t* message2)
{
	mpz_t temp;
	paillier_ciphertext_t* cipher1;
	paillier_ciphertext_t* cipher2;
	paillier_ciphertext_t* result = (paillier_ciphertext_t*)malloc(sizeof(paillier_ciphertext_t));
	
	mpz_init(temp);
	mpz_init(result->c);

	cipher1 = paillier_enc(0, pubkey, message1, paillier_get_rand_devurandom);// Enc(m1)
	cipher2 = paillier_enc(0, pubkey, message2, paillier_get_rand_devurandom);// Enc(m2)
		
	mpz_mul(temp, cipher1->c, cipher2->c); // Enc(m1)*Enc(m2)
	mpz_mod(temp, temp, pubkey->n_squared); // Enc(m1)*Enc(m2) mod n^2
	
	mpz_set(result->c, temp);
	free(cipher1);
	free(cipher2);
	return result;
}

/*
	Apply additive decryption circuit
	Dec(c);
 */
paillier_plaintext_t* 
dec_add_circuit(
	paillier_keys* keys,
	paillier_ciphertext_t* cipher3)
{
	
	paillier_plaintext_t* plaintext;
	plaintext = paillier_dec(0, keys->pubkey, keys->prvkey, cipher3);	

	return plaintext;
}


/*
	Apply multiplicative encryption circuit
	Enc((m1-b1)(m2-b2) mod n) * Enc(b1)^(m2-b2) * Enc(b2)^(m1-b1) mod n^2;
 */
paillier_ciphertext_t* 
enc_mul_circuit(
	paillier_pubkey_t* pubkey, 
	paillier_plaintext_t* message1, 
	paillier_plaintext_t* message2, 
	paillier_ciphertext_t* secret)
{

	gmp_randstate_t rand;
	mpz_t b1, b2, temp1, temp2, temp3;
	paillier_ciphertext_t* cipher1;
	paillier_ciphertext_t* cipher2;
	paillier_ciphertext_t* result = (paillier_ciphertext_t*)malloc(sizeof(paillier_ciphertext_t));
	paillier_plaintext_t* plaintext_temp = (paillier_plaintext_t*)malloc(sizeof(paillier_plaintext_t));
	paillier_ciphertext_t* cipher_temp;
	
	mpz_init(b1);
	mpz_init(b2);
	mpz_init(temp1);
	mpz_init(temp2);
	mpz_init(temp3);
	mpz_init(result->c);

	init_rand(rand, paillier_get_rand_devurandom, pubkey->bits);
	/* ramdomly generate b1 and b2 ranging from (0,n) */
	do
		mpz_urandomb(b1, rand, pubkey->bits); // 0 < b1 (bit length) <= n
	while(mpz_cmp(b1, pubkey->n) >= 0 );
	do
		mpz_urandomb(b2, rand, pubkey->bits); // 0 < b2 (bit length) <= n
	while(mpz_cmp(b2, pubkey->n) >= 0 );
	
	mpz_sub(temp1, message1->m, b1); // m1-b1
	mpz_mod(temp1, temp1, pubkey->n); // m1-b1 mod n
	mpz_sub(temp2, message2->m, b2); // m2-b2
	mpz_mod(temp2, temp2, pubkey->n); // m2-b2 mod n

	/* (m1-b1)(m2-b2) mod n */
	mpz_mul(temp3, temp1, temp2); 
	mpz_mod(temp3, temp3, pubkey->n);  
	
	/* Enc((m1-b1)(m2-b2) mod n) */
	mpz_init(plaintext_temp->m);	
	mpz_set(plaintext_temp->m, temp3);
	cipher_temp = paillier_enc(0, pubkey, plaintext_temp, paillier_get_rand_devurandom);
	mpz_set(temp3, cipher_temp->c);

	mpz_set(plaintext_temp->m, b1);
	cipher1 = paillier_enc(0, pubkey, plaintext_temp, paillier_get_rand_devurandom); // Enc(b1)
	mpz_set(plaintext_temp->m, b2);
	cipher2 = paillier_enc(0, pubkey, plaintext_temp, paillier_get_rand_devurandom); // Enc(b2)
	
	mpz_powm(temp2, cipher1->c, temp2, pubkey->n_squared); // Enc(b1)^(m2-b2) mod n^2
	mpz_powm(temp1, cipher2->c, temp1, pubkey->n_squared); // Enc(b2)^(m1-b1) mod n62

	/* Enc((m1-b1)(m2-b2) mod n) * Enc(b1)^(m2-b2) * Enc(b2)^(m1-b1) mod n^2 */
	mpz_mul(temp3, temp3, temp1); 
	mpz_mul(temp3, temp3, temp2);
	mpz_mod(temp3, temp3, pubkey->n_squared);
	mpz_set(result->c, temp3);

	mpz_mul(temp3, b1, b2); 
	mpz_mod(temp3, temp3, pubkey->n);
	mpz_set(secret->c, temp3);

	return result;
}


/*
	Apply multiplicative decryption circuit
	Dec(c) + secret;
 */
paillier_plaintext_t* 
dec_mul_circuit(
	paillier_keys* keys,
	paillier_ciphertext_t* cipher,
	paillier_ciphertext_t* secret)
{
	mpz_t temp;
	mpz_init(temp);
	paillier_plaintext_t* plaintext;
	paillier_plaintext_t* result = (paillier_plaintext_t*)malloc(sizeof(paillier_plaintext_t));
	mpz_init(result->m);

	plaintext = paillier_dec(0, keys->pubkey, keys->prvkey, cipher);

	mpz_add(temp, plaintext->m, secret->c); // Dec(c)+secret
	mpz_mod(temp, temp, keys->pubkey->n); 
	mpz_set(result->m, temp);
	return result;
}

	
