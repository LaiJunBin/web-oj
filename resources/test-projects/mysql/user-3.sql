
select lastName, firstName, city
from person inner join address
on person.personId = address.personId;