Authorisation is the process of protected resources  from unspecified users
eg: /total-sales should be only showed to admin
this applies mainly for role based access
But there are some endpoints that one user can use to access others data
for that we need to add check of for the userId
Eg: GET /users/:id only that specific user should be able to get his data
this can be solved by checking if the id sent by user is same as the ont who send it

OR just use jwt as a source of truth for userId
no relying on user sent params