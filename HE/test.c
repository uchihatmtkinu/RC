#include <stdio.h>
#include <stdlib.h>
#include "paillier_mh.h"

void main(){
	char* message1_str = "This is message 1";
	char* message2_str = "This is message 2";
	paillier_keys* keys = (paillier_keys*)malloc(sizeof(paillier_keys));
	paillier_plaintext_t* message1 = (paillier_plaintext_t*)malloc(sizeof(paillier_plaintext_t));
	paillier_plaintext_t* message2 = (paillier_plaintext_t*)malloc(sizeof(paillier_plaintext_t));
	paillier_ciphertext_t* cipher;
	paillier_ciphertext_t* secret = (paillier_ciphertext_t*)malloc(sizeof(paillier_ciphertext_t));
	paillier_plaintext_t* result;

	mpz_t temp;
	mpz_init(temp);
	mpz_init(secret->c);
	
	printf("Generating keys for paillier cryptosystem...");
	key_gen(keys);
	printf("Done\n");

	printf("Message 1 string: \"%s\"\nMessage 2 string: \"%s\"\n", message1_str, message2_str);
	message1 = paillier_plaintext_from_str(message1_str);
	message2 = paillier_plaintext_from_str(message2_str);
	
	/* test additive homomorphic */
	printf("\nProving Additive Homomorphic: \n");
	mpz_add(temp, message1->m, message2->m); // m1+m2
	mpz_mod(temp, temp, keys->pubkey->n); // m1+m2 mod n
	gmp_printf("m1+m2 mod n \t\t= %Zd\n", temp);

	cipher = enc_add_circuit(keys->pubkey, message1, message2);// c = Enc(m1) + Enc(m2)
	result = dec_add_circuit(keys, cipher);// Dec(c)
	gmp_printf("Dec_c(Enc_c(m1, m2)) \t= %Zd\n", result->m);
		
	
	/* test multiplicative homomorphic */
	printf("\nProving Multiplicative Homomorphic:\n");
	mpz_mul(temp, message1->m, message2->m); // m1*m2
	mpz_mod(temp, temp, keys->pubkey->n); // m1*m2 mod n
	gmp_printf("m1*m2 mod n \t\t= %Zd\n", temp);

		
	cipher = enc_mul_circuit(keys->pubkey, message1, message2, secret);// c = Enc(m1) + Enc(m2)
	result = dec_mul_circuit(keys, cipher, secret);// Dec(c)
	//gmp_printf("secret: %Zd\n", secret->c);
	gmp_printf("Dec_c(Enc_c(m1, m2)) \t= %Zd\n", result->m);

	free(keys->pubkey);
	free(keys->prvkey);
	free(message1);
	free(message2);
	free(keys);
}
