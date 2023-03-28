# Microservices Architecture for Beginners

This short document contains my view about microservices, based on various resources I found online. At the moment of writing this, microservices is a hot topic, especially found in Go applications, and the motivation behind writing this document was that after over 1 year of developing microservices, I realized that I wouldn't really be able to tell why we're using them. Of course, there's the classical "it's easier to scale" and "it's easier to develop" type of arguments, but here I will be focusing not on _why_ these points are valid, because they're not universally true, thus rather _when_ they're valid. I will also attempt to explain the difference between a microservices architecture and a monolith architecture, while providing a short proof-of-concept example.
 
> **NB:** As a disclaimer, the document here is just my personal opinion that I've formed on the topic, by working with both microservices and monoliths, reading resources, watching videos, talking to other people etc. I'm open to update the article based on feedback. Resources I've consulted are linked at the bottom of the document, and referenced in text.

So what's a monolith and what's a microservice?
-----------------------------------------------

Generally, I tend to view monoliths as nothing more than very big programs, and microservices as smaller programs running together. A good question to ask here would be, how big is a big program? What's the line that separates a monolith from a smaller microservice? Unfortunately, there is no clear answer to these questions because it largely depends on context. However, a general consensus would be that if the entire application runs as a single unit of execution (so you run it entirely at once) it's a monolith. If running the application consists of running individual apps separately, then we have microservices.

This description is cutting some corners, only providing an intuitive overview of the difference between the architectures. The term "app" or "application" is intentionally very broad, because it's not easy to put together a definition for the pieces that constitute a software application. [[1]](https://www.pcmag.com/encyclopedia/term/application-software) Take for example an application presenting a lighter separation: a frontend client running separately and communicating with a backend, that is connected to a database management system. Would this be considered a microservices application, just because the backend, frontend and the database management system are running in separate units of execution? 

As described by [microservices.io](https://microservices.io/) [[2]](https://microservices.io/), in order for an application's architecture to be considered microservices, it should come with independently deployable modules, that are loosely coupled and modeled around the bussiness requirements of the application. The modules should be developed and scaled individually, often by separate teams of the organization. With this definition in mind, I would consider the example above to classify as a microservices architecture, albeit a rather introductory one which is not quite what the term "microservices" is generally used for. The bussiness separation here (and thus the microservices) would be the user experience, taken care by the frontend module, and the server supplying the data, comprised of the backend and the database management server. Do note that I intentionally included the database management server as part of the service; I will come back to this later in the article.

Perhaps not as obvious, but running multiple instances of the same monolith application, most often with a load balancer to even out the processing toll on each instance, doesn't classify as a microservice, but rather as horizontal scaling. The core logic is executed inside a single instance, and not spread accross multiple ones. Therefore, when you make a request to this wall of applications, the load balancer routes it to a single instance which processes it, without communicating with the other instances (note that communication with other shared processes such as database systems may still be involved, though). With a microservices architecture, processing a request could involve traversing multiple of these "small" applications before returning a result, as we have with the [API Gateway Pattern.](https://learn.microsoft.com/en-us/dotnet/architecture/microservices/architect-microservice-container-applications/direct-client-to-microservice-communication-versus-the-api-gateway-pattern) [[3]](https://learn.microsoft.com/en-us/dotnet/architecture/microservices/architect-microservice-container-applications/direct-client-to-microservice-communication-versus-the-api-gateway-pattern)

I'll further explain this point by providing the example from [TechWorld with Nana's](https://www.youtube.com/@TechWorldwithNana) YouTube video ["Microservices explained - the What, Why and How?"](youtube.com/watch?v=rv4LlmLmVWk) [[4]](youtube.com/watch?v=rv4LlmLmVWk). Given an online shop application, you might find multiple components: user authentication, poduct catalogue, shopping cart, notification system and a payment system. If there is a single codebase and starting the application implies running all these components together as a single unit, it's considered to be a monolith application. If you run each of these components separately in a smaller application and have them communicate with each other over network calls by defining an interface for each of them, that's indicative of a microservices architecture.

 
Note that it's still perfectly valid to have separate modules and separate teams working on them in a monolithic application. The section [When to use microservices](#when-to-use-microservices) discusses in more detail the difference between microservices and modules, and when to consider using each one.

Monoliths vs Mircoservices
--------------------------

In this next section, I will explain the fundamental differences between the two architectures. This is an objective comparison, and doesn't attempt to glamorize one architecture over the other, but rather to illustrate which development areas might fit better with one architecture. The ideal architecture of an application is largely dependent on the bussiness needs and the scale it is trying to reach, and coming up with a one-size-fits-all formula to determine the ideal choice for every project is not possible.

That being said, let's first resume a previous point, that a monolith runs as a single unit of execution, whereas microservices run as separate independent modules. This property has direct consequences on the development process and the technology choice. Usually, monoliths are written with a single technology stack, often even with the same programming language. [[5]](https://ginbits.com/monolithic-microservices-architecture-a-harmonized-dance/) [[6]](https://microservices.io/patterns/monolithic.html) In contrast, microservices are smaller applications that can run independently, exposing certain interfaces for communication. It is not a hard requirement to write the services using the same technology, e.g. one service may be written in Python, whereas another service may be written in Go.

It is worth mentioning though that sometimes programs are developed with multiple languages, even when they run as a single process. [[7]](https://stackoverflow.com/questions/636841/how-do-multiple-languages-interact-in-one-project)

I mentioned previously that there are consequences affecting the development process. In monoliths, the effects of modifying one part of the application propagate throughout the entire application. When code changes, it is possible that other parts of the application break, if they're not compatible with the changes. This is less noticeable within a single microservices, where there is generally less code in each microservice, but still possible nonetheless. However, this doesn't mean that a change to a single microservice never affects the others: updates to the communication interfaces may incur changes to the service communication, be it with RPC calls, HTTP requests etc. For example, if one service exposes a gRPC server, removing an RPC method requires changes to all other services which were using the method.

In terms of external code, dependency management may be less flexible with monoliths, particularly because it's not always possible to support multiple versions of the same dependency, or at least not easily, if that is desired. [[8]](https://boyl.es/post/two-versions-same-library/) With microservices though, dependencies are not shared, but this also means that each microservice has to import its own dependencies, often leading to having to repeat the same dependencies. Although, in practice, this is usually not a problem, since often languages offer ways to handle external modules easily, e.g. via [go modules](https://go.dev/blog/using-go-modules) in Go, and it's often the case that microservices are developed by separate teams, which is what we will talk about shortly below.

Consequently, it's also inherently more difficult to share and reuse code in a microservices architecture, when in monoliths it's mostly just a matter of following the [SOLID](https://www.baeldung.com/solid-principles) principles in order to write more maintainable and reusable code. Shared functionality such as database pagination, graceful server shutdown or testing setups may be written and packaged in a common module, but this requires additional work to make the package available, and doesn't play well if using different programming languages for the microservices. However, the process of testing the shared code or modifying it should have similar implications in both architectures.

The building and deployment process is also different, since monoliths are generally built and deployed as a single package. Changes to any part of the code in a monolithic application requires re-building and re-packaging the application, which in turn could affect the development process, especially when the application grows large enough that it takes a lot of time to compile it. On the other hand, while the microservices are separately built and packaged themselves, the process is individual for each one: deploying changes to one microservice doesn't affect the other microservices, making it possible to apply changes even while the other microservices are running (unless there are breaking changes to the communication interfaces, as explained in a previous paragraph).

Indirectly, all of the above may have an impact on the communication between teams. Software is generally developed in teams, and teams often work on different parts of the application. This can be seen with both architectures: in monoliths, teams often work on different modules, and with microservices, teams often work on individual services. Monoliths tend to require more synchronization between teams before pushing any new changes, to minimize the risk of potential side effects, such as breaking code changes or dependency conflicts.

Not only the communication between teams is different, but also boarding new team members brings different challenges. As mentioned in [TechWorld with Nana's](https://www.youtube.com/@TechWorldwithNana), as well as [Gaurav Sen's](https://www.youtube.com/@gkcs) video ["What is a microservice architecture and it's advantages?"](https://www.youtube.com/watch?v=qYhRvH9tJKw), [[9]] a monolith architecture often requires all team members to have some knowlledge about the entire system, in order to minimize the risk of side effects and improving development efficiency. This is not necessarily the case with microservices, where if working on a single microservice it's possible to treat the others as "black boxes" and know only about the communication interfaces and the computations to be performed for each operation. Of course, this mindset may be applied to monoliths as well, e.g. you might only take a look at the functions exposed by a module and their documentation, similar to [C++ header files](https://learn.microsoft.com/en-us/cpp/cpp/header-files-cpp?view=msvc-170), but microservices tend to force it, e.g. team members might not even have access to the code of other services. This shows an additional contrast between the two: it's easier to restrict developer access to the internal code of other microservices.


-- single point of failure

CI/CD

The table below illustrates a short summary of the paragraph.

| Monoliths                                 | Microservies                                 |
| ----------------------------------------- | -------------------------------------------- |
| Developed, deployed and scaled as a unit; | Developed, deployed and scaled individually; |
|                                           |                                              |





This section describes a short summary of the difference between the two architectures that we discused previously.

-- different teams
-- changing code

Now that this is out of the way, 


Microservices are loosely coupled, but the code of a monolith may also be loosely coupled. The difference is that these can be worked on and developed individually without affecting the others, as long as the interface they provide stays the same. Dependencies are also managed individually for each service and do not interfere with other parts of the project. Code might also be loosely coupled, but subtle things like changing a dependency might affect the integrity of the monolith and damage other parts silently, whereas this doesn't happen with the microservices. They don't _encourage_ you to decouple your code, they _force_ you too. Of course, this is only useful as long as it actually makes sense to decouple the code. In any case, I will however take this as a PRO for microservices.

That doesn't mean there aren't CONS, though. Running decoupled code in separate programs also means that there's no more "direct" communication, as it would have been in a monolith. It's also generally more difficult to implement those interfaces, since you're not just calling a function, there's already inter-process communication involved. Moreover, if the microservices are not running in the same network, they also become susceptible to the "fallacies of distributed systems" (https://architecturenotes.co/fallacies-of-distributed-systems/). However, it's usually the case that they run at least on the same netowrk if not on the same computing instance*, and we do have tools that greatly facilitate the communication while also keeping it efficient such as gRPC, so this is not a real drawback that stands in the way. Nevertheless, there's the additional work of implementing the communication, and the execution IS slower, so one should be careful to not oversplit. I usually use my own judgement and go for the largest unit that makes sense as a standalone application.

(it doesn't make sense to have the microservices constituting application A run on separate networks just for the sake of it -- monolith applications are also running on distributed computing instances)

When to use microservices?
-------------------------

covers modules

- deciding on their size


Think of an app that amongst allows you to upload files:

Resources
---------

1. https://www.pcmag.com/encyclopedia/term/application-software
2. https://microservices.io/
3. https://learn.microsoft.com/en-us/dotnet/architecture/microservices/architect-microservice-container-applications/direct-client-to-microservice-communication-versus-the-api-gateway-pattern
4. youtube.com/watch?v=rv4LlmLmVWk
5. https://ginbits.com/monolithic-microservices-architecture-a-harmonized-dance/
6. https://microservices.io/patterns/monolithic.html
7. https://stackoverflow.com/questions/636841/how-do-multiple-languages-interact-in-one-project
8. https://boyl.es/post/two-versions-same-library/
9. https://www.youtube.com/watch?v=qYhRvH9tJKw
https://architecturenotes.co/fallacies-of-distributed-systems/