# Web Archive

A java web application is delivered as a war archive. This archive contains some jar archives taken from the JVM environment. Other archives, like the servlet logic, JSP, JSF etc. are delivered by the container (Tomcat, Jetty) or Application Server (GlassFish, Jboss/WildFly). 

Other dependencies specific to the application can be located elsewhere. The directory WEB-INF/lib holds all these dependencies.

1. Build the entire project using the command 

```
gradle clean build
```

The build output is similar to what we've seen previously with jar dependencies. Except this time, a .war archive was generated. The war archive contains both a `WEB-INF` and a `META-INF` directory.

Inside `META-INF` there's the usual manifest (just with the version), and inside `WEB-INF` there's a `classes/` directory, with the generated Hw binary.

The `.war` archive can be deployed on a container or an application server, which can be executed and accessed from the browser.