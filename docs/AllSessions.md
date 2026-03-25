Group all the user login and store in a redis set through SAdd
the key will be user_session:userid -> hash1, hash2, hash3
this will alsi help in all sessions logout