# f5-api-task

## how to run:
- in command line run `docker-compose up --build`
- alternatively run `go run main.go`

## changes I made:

1. in `Register` function, I added an input validation to make sure that the provided role is admin/user only
2. in `AccountsHandler`, I moved `if claims.Role != "admin" ` to the top of the function because both operations are allowd to admins only, and `listAccounts` didnt enforce it.
3. in `AccountsHandler`, I added a "method not allowed" error in case client sent an unsuported method.
4. in `BalanceHandler`, I added a check that the clients role is "user" (unless its a get request because an admin can view all of the users balance).
5. in `getBalance`,  `depositBalance`, `withdrawBalance`, I added a check to see if the user_id provided matches the user from the claims, to prevent BOLA. in `getBalance` I also check if the user is not an admin because admins can view all of the users balances.
before my change any user could pass any user_id he wanted and get the data of different users.
I also added an input validation to make sure the amount sent by the user is positive.

## extra suggestions:
- In the api all IDs are created by incrementation. Defining IDs this way is not safe, it makes it easy for someone to guess IDs. I would define user id as a UUID, so if an attacker will find a vaulnerability (BOLA for example), he wont be able to guess the resoueces IDs. I didnt implement the change to UUID because in the api_usage_example i recived, the response example contains the id as int, and I wanted to follow it exactly.
- saving raw passwords is bad security wise. I would change the way i store the password to storing a hash of the password. I also didnt implement it because in the api_usage_example i recived the response example contains the raw password, and I wanted to follow it exactly.
