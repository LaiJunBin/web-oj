
select firstName, state
from person left join address
on person.personId = address.personId
where state IS NOT NULL;