# Ultimate Software Design with Kubernetes

[![CircleCI](https://circleci.com/gh/ardanlabs/service.svg?style=svg)](https://circleci.com/gh/ardanlabs/service)
[![Go Report Card](https://goreportcard.com/badge/github.com/ardanlabs/service)](https://goreportcard.com/report/github.com/ardanlabs/service)
[![go.mod Go version](https://img.shields.io/github/go-mod/go-version/ardanlabs/service)](https://github.com/ardanlabs/service)

[![InsightsSnapshot](https://dl.circleci.com/insights-snapshot/gh/ardanlabs/service/master/workflow/badge.svg)](https://app.circleci.com/insights/github/ardanlabs/service/workflows/workflow/overview?branch=master)

Copyright 2018, 2019, 2020, 2021, 2022, 2023, 2024 Ardan Labs  
hello@ardanlabs.com

## My Information

```
Name:    Bill Kennedy  
Company: Ardan Labs  
Title:   Managing Partner  
Email:   bill@ardanlabs.com  
Twitter: goinggodotnet  
```

## Description

_"As a program evolves and acquires more features, it becomes complicated, with subtle dependencies between components. Over time, complexity accumulates, and it becomes harder and harder for programmers to keep all the relevant factors in their minds as they modify the system. This slows down development and leads to bugs, which slow development even more and add to its cost. Complexity increases inevitably over the life of any program. The larger the program, and the more people that work on it, the more difficult it is to manage complexity."_ - John Ousterhout  

The service starter kit is a starting point for building production grade scalable web service applications that leverage the power of a Domain Driven, Data Oriented Architecture that can run in Kubernetes. The goal of this project is to provide a proven starting point that reduces the repetitive tasks required for a new project to be launched into production. It uses minimal dependencies, implements idiomatic code and follows Go best practices. Collectively, the project lays out everything logically to minimize guess work and enable engineers to quickly maintain a mental model for the project.

The class behind this starter kit teaches how to build production-level software in Go leveraging the power of a Domain Driven, Data Oriented Architecture that can run in Kubernetes. From the beginning, you will pair program with the instructor walking through the design philosophies and guidelines for building software in Go. With each new feature that is added to the project, you will learn how to deploy to and manage the Kubernetes environment used to run the project. The core of this class is to teach you and your team how to handle and reduce the spread of complexity in the systems you are building.

Learn more about the project:

[Wiki](https://github.com/ardanlabs/service/wiki) | [Course Outline](https://github.com/ardanlabs/service/wiki/course-outline) | [Class Schedule](https://www.ardanlabs.com/events/)

## Licensing

```
Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
```

## Learn More

**Reach out about corporate training events, open enrollment live training sessions, and on-demand learning options.**

Ardan Labs (www.ardanlabs.com)  
hello@ardanlabs.com

## Index

* [Installation](https://github.com/ardanlabs/service?tab=readme-ov-file#installation)
* [Create Your Own Version](https://github.com/ardanlabs/service?tab=readme-ov-file#create-your-own-version)
* [Running The Project](https://github.com/ardanlabs/service?tab=readme-ov-file#running-the-project)
* [Purchase Video](https://github.com/ardanlabs/service?tab=readme-ov-file#purchase-video)
* [Experience](https://github.com/ardanlabs/service?tab=readme-ov-file#our-experience)
* [Teacher](https://github.com/ardanlabs/service?tab=readme-ov-file#our-teacher)
* [More About Go](https://github.com/ardanlabs/service?tab=readme-ov-file#more-about-go)
* [Minimal Qualified Student](https://github.com/ardanlabs/service?tab=readme-ov-file#minimal-qualified-student)
* [Joining the Go Slack Community](https://github.com/ardanlabs/service?tab=readme-ov-file#joining-the-go-slack-community)

## Installation

To clone the project, create a folder and use the git clone command. Then please read the [makefile](makefile) file to learn how to install all the tooling and docker images.

```
$ cd $HOME
$ mkdir code
$ cd code
$ git clone https://github.com/ardanlabs/service or git@github.com:ardanlabs/service.git
$ cd service
```

## Create Your Own Version

If you want to create a version of the project for your own use, use the new gonew command.

```
$ go install golang.org/x/tools/programs/gonew@latest

$ cd $HOME
$ mkdir code
$ cd code
$ gonew github.com/ardanlabs/service github.com/mydomain/myproject
$ cd myproject
$ go mod vendor
```

Now you have a copy with your own module name. Now all you need to do is initialize the project for git.

## Running The Project

To run the project use the following commands.

```
# Install Tooling
$ make dev-gotooling
$ make dev-brew
$ make dev-docker

# Run Tests
$ make test

# Run Project
$ make dev-up
$ make dev-udpate-apply
$ make token
$ export TOKEN=<COPY TOKEN>
$ make users

# Run Load
$ make load

# Run Tooling
$ make grafana
$ make statsviz

# Shut everything down
$ make dev-down
```

## Purchase Video

The entire training class has been recorded to be made available to those who can't have the class taught at their company or who can't attend a conference. This is the entire class material.

[ardanlabs.com/education](https://www.ardanlabs.com/education/)

## Our Experience

We have taught Go to thousands of developers all around the world since 2014. There is no other company that has been doing it longer and our material has proven to help jump-start developers 6 to 12 months ahead of their knowledge of Go. We know what knowledge developers need in order to be productive and efficient when writing software in Go.

Our classes are perfect for intermediate-level developers who have at least a few months to years of experience writing code in Go. Our classes provide a very deep knowledge of the programming langauge with a big push on language mechanics, design philosophies and guidelines. We focus on teaching how to write code with a priority on consistency, integrity, readability and simplicity. We cover a lot about “if performance matters” with a focus on mechanical sympathy, data oriented design, decoupling and writing/debugging production software.

## Our Teacher

### William Kennedy ([@goinggodotnet](https://twitter.com/goinggodotnet))  
_William Kennedy is a managing partner at Ardan Labs in Miami, Florida. Ardan Labs is a high-performance development and training firm working with startups and fortune 500 companies. He is also a co-author of the book Go in Action, the author of the blog GoingGo.Net, and a founding member of GoBridge which is working to increase Go adoption through diversity._

_**Video Training**_  
[Ultimate Go Video](https://education.ardanlabs.com)  
[Ardan Labs YouTube Channel](http://youtube.ardanlabs.com/)

_**Blog**_  
[Going Go](https://www.ardanlabs.com/blog/)    

_**Writing**_  
[Running MongoDB Queries Concurrently With Go](http://blog.mongodb.org/post/80579086742/running-mongodb-queries-concurrently-with-go)    
[Go In Action](https://www.manning.com/books/go-in-action)  

_**Articles**_  
[IT World Canada](http://www.itworldcanada.com/article/nascent-google-development-language-shows-promise-for-more-productive-coding/387449)

_**Video**_  
[Golang Charlotte (2024) - Domain Driven, Data Oriented Architecture](https://www.youtube.com/watch?v=bQgNYK1Z5ho)  
[GopherCon SG (2023) - K8s CPU Limits and Go](https://www.youtube.com/watch?v=Dm7yuoYTx54&list=PLq2Nv-Sh8Eba2gEaId35K2aAUFdpbKx9D&index=6)  
[P99 Talk (2022) - Evaluating Performance In Go](https://www.youtube.com/watch?v=PYMs-urosXs)  
[GopherCon Europe (2022) - Practical Memory Profiling](https://www.youtube.com/watch?v=6qAfkJGWsns)  
[Dgrpah Day (2021) - Getting Started With Dgraph and GraphQL](https://www.youtube.com/watch?v=5L4PUbDqSEo)  
[GDN Event #1 (2021) - GoBridge Needs Your Help](https://www.youtube.com/watch?v=Tst0oI97cvQ&t=2s)  
[Training Within The Go Community (2019)](https://www.youtube.com/watch?v=PSR1twjzzAM&feature=youtu.be)  
[GopherCon Australia (2019) - Modules](https://www.youtube.com/watch?v=MVxbVR_6Tac)  
[Golab (2019) - You Want To Build a Web Service?](https://www.youtube.com/watch?v=IV0wrVb31Pg)  
[GopherCon Singapore (2019) - Garbage Collection Semantics](https://www.youtube.com/watch?v=q4HoWwdZUHs)  
[GopherCon India (2019) - Channel Semantics](https://www.youtube.com/watch?v=AHAf1Xfr_HE)  
[GoWayFest Minsk (2018) - Profiling Web Apps](https://www.youtube.com/watch?v=-GBMFPegqgw)  
[GopherCon Singapore (2018) - Optimizing For Correctness](https://engineers.sg/video/optimize-for-correctness-gopherconsg-2018--2610)  
[GopherCon India (2018) - What is the Legacy You Are Leaving Behind](https://www.youtube.com/watch?v=j3zCUc06OXo&t=0s&index=11&list=PLhJxE57Cki63cElK2kmt3_vi8j2eIHTqZ)  
[Code::Dive (2017) - Optimizing For Correctness](https://www.youtube.com/watch?v=OTLjN8NQDyo)  
[Code::Dive (2017) - Go: Concurrency Design](https://www.youtube.com/watch?v=OrctYMf4btA)  
[dotGo (2017) - Behavior Of Channels](https://www.youtube.com/watch?v=zDCKZn4-dck)  
[GopherCon Singapore (2017) - Escape Analysis](https://engineers.sg/video/escape-analysis-and-memory-profiling-gophercon-sg-2017--1746)  
[Capital Go (2017) - Concurrency Design](https://www.youtube.com/watch?v=yGOOUCrrgrE&index=10&list=PLeGxIOPLk9EKdl-h_Y-sbLhLoP-ia7CJ5)  
[GopherCon India (2017) - Package Oriented Design](https://www.youtube.com/watch?v=spKM5CyBwJA#t=0m56s)  
[GopherCon India (2015) - Go In Action](https://www.youtube.com/watch?v=QkPw8-Pf0SM)  
[GolangUK (2016) - Dependency Management](https://youtu.be/CdhucJShJU8)  
[GothamGo (2015) - Error Handling in Go](https://vimeo.com/115782573)  
[GopherCon (2014) - Building an analytics engine](https://www.youtube.com/watch?v=EfJRQ1lGkUk)  

[Golang Charlotte (2023) - Domain Driven, Data Oriented Architecture with Bill Kennedy](https://www.youtube.com/watch?v=bQgNYK1Z5ho)  
[Prague Meetup (2021) - Go Module Engineering Decisions](https://youtu.be/m8lgcXv2lhI)  
[Practical Understanding Of Scheduler Semantics (2021)](https://www.youtube.com/watch?v=p2Cjq3Dq2Q0)  
[Go Generics Draft Proposal (2020)](https://www.youtube.com/watch?v=gIEPspmbMHM&t=2069s)  
[Hack Potsdam (2017) - Tech Talk with William Kennedy](https://www.youtube.com/watch?v=sBzJ-sjhgs8)  
[Chicago Meetup (2016) - An Evening](https://vimeo.com/199832344)  
[Vancouver Meetup (2016) - Go Talk & Ask Me Anything With William Kennedy](https://www.youtube.com/watch?v=7YcLIbG1ekM&t=91s)  
[Vancouver Meetup (2015) - Compiler Optimizations in Go](https://www.youtube.com/watch?v=AQipeq39Aek)  
[Bangalore Meetup (2015) - OOP in Go](https://youtu.be/gRpUfjTwSOo)  
[GoSF Meetup - The Nature of Constants in Go](https://www.youtube.com/watch?v=ZUCHMAoOgUQ)    
[London Meetup - Mechanical Sympathy](https://skillsmatter.com/skillscasts/8353-london-go-usergroup)    
[Vancouver Meetup - Decoupling From Change](https://www.youtube.com/watch?v=7YcLIbG1ekM&feature=youtu.be)  

_**Podcasts**_  
[Ardan Labs Podcast: On Going Series](https://ardanlabs.buzzsprout.com/)  
[Encore, domain design in Go with Bill Kennedy](https://gopodcast.dev/episodes/034-encore-domain-design-in-go-with-bill-kennedy)  
[Mangtas Nation: A Golang Deep Dive with Bill Kennedy](https://anchor.fm/mangtasnation/episodes/A-Golang-Deep-Dive-with-Bill-Kennedy--S2-EP3-e1ij9c3)  
[Coding with Holger: Go with Bill Kennedy](https://anchor.fm/coding-with-holger/episodes/Go-with-Bill-Kennedy-e1c9h2q)  
[Craft of Code: From Programming to Teaching Code with Bill Kennedy](https://podcasts.apple.com/us/podcast/from-programming-to-teaching-code-with-bill-kennedy/id1537136353?i=1000545230339)  
[GoTime: Design Philosophy](https://changelog.com/gotime/172)  
[GoTime: Learning and Teaching Go](https://changelog.com/gotime/72)  
[GoTime: Bill Kennedy on Mechanical Sympathy](https://changelog.com/gotime/6)  
[GoTime: Discussing Imposter Syndrome](https://changelog.com/gotime/30)  
[HelloTechPros: Your Tech Interviews are Scaring Away Brilliant People](http://hellotechpros.com/william-kennedy-people)    
[HelloTechPros: The 4 Cornerstones of Writing Software](http://hellotechpros.com/bill-kennedy-productivity)  

## More About Go

Go is an open source programming language that makes it easy to build simple, reliable, and efficient software. Although it borrows ideas from existing languages, it has a unique and simple nature that make Go programs different in character from programs written in other languages. It balances the capabilities of a low-level systems language with some high-level features you see in modern languages today. This creates a programming environment that allows you to be incredibly productive, performant and fully in control; in Go, you can write less code and do so much more.

Go is the fusion of performance and productivity wrapped in a language that software developers can learn, use and understand. Go is not C, yet we have many of the benefits of C with the benefits of higher level programming languages.

[The Ecosystem of the Go Programming Language](https://henvic.dev/posts/go/) - Henrique Vicente  
[The Why of Go](https://www.infoq.com/presentations/go-concurrency-gc) - Carmen Andoh  
[Go Ten Years and Climbing](https://commandcenter.blogspot.com/2017/09/go-ten-years-and-climbing.html) - Rob Pike  
[The eigenvector of "Why we moved from language X to language Y"](https://erikbern.com/2017/03/15/the-eigenvector-of-why-we-moved-from-language-x-to-language-y.html) - Erik Bernhardsson  
[Learn More](https://talks.golang.org/2012/splash.article) - Go Team  
[Simplicity is Complicated](https://www.youtube.com/watch?v=rFejpH_tAHM) - Rob Pike  
[Getting Started In Go](http://aarti.github.io/2016/08/13/getting-started-in-go) - Aarti Parikh  

## Minimal Qualified Student

The material has been designed to be taught in a classroom environment. The code is well commented but missing some contextual concepts and ideas that will be covered in class. Students with the following minimal background will get the most out of the class.

* Studied CS in school or has a minimum of two years of experience programming full time professionally.
* Familiar with structural and object oriented programming styles.
* Has worked with arrays, lists, queues and stacks.
* Understands processes, threads and synchronization at a high level.
* Operating Systems
	* Has worked with a command shell.
	* Knows how to maneuver around the file system.
	* Understands what environment variables are.

## Joining the Go Slack Community

We use a Slack channel to share links, code, and examples during the training.  This is free.  This is also the same Slack community you will use after training to ask for help and interact with may Go experts around the world in the community.

1. Using the following link, fill out your name and email address: https://invite.slack.gobridge.org
1. Check your email, and follow the link to the slack application.
1. Join the training channel by clicking on this link: https://gophers.slack.com/messages/training/
1. Click the “Join Channel” button at the bottom of the screen.
___
All material is licensed under the [Apache License Version 2.0, January 2004](http://www.apache.org/licenses/LICENSE-2.0).
