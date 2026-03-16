Salting is the process of adding the random value (salt) to a password before hashing.
Randomly generated on every password so that attackers can't use rainbow table.(precomputed hash table)


hash = Hash(password + salt)
even with same password hash will be different.
Rainbow tables rely on precomputed hashes of common passwords.
So precomputed passwords won't work since even for the same password hash would be completely different.

How does Comparing work?
You have user with name:god_slayer password:password123 and created salt:1234wqq

and stored hash as 1234wqq$hashStored

get salt and hash -> 1234wqq and hashStored

now run the algo for 1234wqq + password123 -> newhash 

if the newHash == hashStored
    then same user