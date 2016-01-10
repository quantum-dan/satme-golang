# satme
SATme platform.  This version is written in Golang and currently not being used for production.

## Why Golang?
Originally, I'd been planning to use Haskell with the Yesod web framework, and I still have this in mind for the future.  However, Haskell is too complex to achieve proficiency in the short timeframe we're aiming for before initial launch.

Golang, then, becomes an excellent second choice.  Its green threads do well with concurrency, like Node.js and Haskell, and it is reasonably performant in general as compared to Python, Node, etc.  In addition, it's easy to learn, easy to use, has an excellent and simple web toolkit in the form of Gorilla, and has the advantages of being statically-typed.

A further advantage, particularly as compared to Haskell, is that I can easily teach it to the software engineering club from which I'll be drawing my first engineers (we can't afford professionals yet: the one writing this is the closest thing we've got, and I'm working for free).

## Technologies
* Gorilla toolkit where applicable (routing, sessions; websockets if we choose to use them)
* MongoDB with mgo driver
* html/templates for the HTML

The development environment in use is Golang 1.5 on FreeBSD/amd64.

## Roadmap
1. Basics (routing, sessions, form handling) [DONE]
2. Database for logins [DONE]
3. Quiz scores tracking & scholarship matching
4. School accounts; student accounts associated with school accounts optionally
5. Counselor tracking of student progress
6. Addition of MentorNet project (not yet started, another idea of mine) for extra-curricular education.
7. 

## Documentation for Libraries
Documentation for the libraries used:
* bcrypt (password hashing): https://godoc.org/golang.org/x/crypto/bcrypt
* Gorilla (web toolkit): http://www.gorillatoolkit.org/
* mgo (MongoDB driver): https://labix.org/mgo
