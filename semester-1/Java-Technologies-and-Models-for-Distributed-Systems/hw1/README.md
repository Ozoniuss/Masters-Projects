# Servlet homework

## Prerequisites

The following environment variables are set:

- `CATALINA_HOME` points to the Tomcat directory;
- `JETTY_HOME` points to the Jetty directory (other versions of jetty might deposit deployments in another directory specified with `JETTY_BASE`);
- `JBOSS_HOME` points to the WildFly directory;
- `GF_HOME` points to the GlassFish directory.

## Technologies and installation

The application is written in Java and uses Java Development Kit (JDK) 8, and is based on Servlets (version 4) to interact with web clients. These servlets are managed by an application servers, which also provides web support.  We presented deploying the application using the Tomcat (tomcat-9.0.69) and Jetty (jetty-9.4.49) servlet containers, as well as the Wildfly (widlfly-21.0.0) and Glassfish (glassfish-5) EJB containers.

## Running the application

There are multiple ways to run the application, and the process is slightly different for each application server. As the first method for all four of them, we've deployed the application by copying its generated `.war` archive into a dedicated directory of every application server. Additionally for Tomcat and Jetty, the app was also deployed by specifying an external context instead of copying the entire archive, as well as coding the server functionality in the application itself and just running the application (embedded).

Next we will cover how to set up each of these deployment methods.

### Copying the war archive

The first step is to build the application using the following command (from the `homework/` directory):

```
gradle clean build
```

This removes old artifacts generated by previous builds and creates a fresh new build. The description of this build is found in the `build.gradle` file, which applies the war plugin to generate a war archive and retrieves the servlet dependency from the Maven central repository. The generated war archive can be found at `build/libs/homework.war`.

For Tomcat, copy the war archive inside the `$CATALINA_HOME.webapps` directory.

```
cp build/libs/homework.war $CATALINA_HOME/webapps
copy build\libs\homework.war "%CATALINA_HOME%\webapps" # Windows
```

Next, start the Tomcat contaier. This can be done with

```
$CATALINA_HOME/bin/startup.sh
%CATALINA_HOME%\bin\startup.bat # Windows
```

Open the browser and access the application at `http://localhost:8080/homework/home`.

In order to stop the Tomcat container, use the command:

```
$CATALINA_HOME/bin/shutdown.sh
%CATALINA_HOME%\bin\shutdown.bat # Windows
```

</br>

For Jetty, the process is similar. Copy the generated war archive into Jetty's home directory:

```
cp build/libs/homework.war $JETTY_HOME/webapps
copy build\libs\homework.war "%JETTY_HOME%\webapps" # Windows
```

Note that in other versions of Jetty, the webapps folder might be located inside a directory called jetty-base.

Next, start the Jetty container. Run this command from the `$JETTY_HOME` directory (TODO: for some reason, doesn't work outside):

```
java -jar start.jar
java -jar start.jar
```

Open the browser and access the application at `http://localhost:8080/homework/home`. To stop the container, just kill its running process in the terminal.

</br>

The same goes for Wildfly. Copy the generated archive inside the `standalone/deployments` folder of the wildfly directory:

```
cp build/libs/homework.war $JBOSS_HOME/standalone/deployments
copy build\libs\homework.war "%JBOSS_HOME%\standalone\deployments" # Windows
```

Start the Wildfly container:

```
$JBOSS_HOME/bin/standalone.sh
%JBOSS_HOME%\bin\standalone.bat # Windows
```

Like Jetty, stop the container by killing its running process in the terminal.

</br>

Finally, the last application server used is GlassFish. Start the server with the command

```
$GF_HOME/bin/asadmin start-domain
%GF_HOME%\bin\asadmin start-domain # Windows
```

Next, deploy the application by copying the war archive into the autodeploy directory from glassfish:

```
cp build/libs/homework.war $GF_HOME/glassfish/domains/domain1/autodeploy
copy build\libs\homework.war "%GF_HOME%\glassfish\domains\domain1\autodeploy" # Windows
```

Finally, once done stop the GlassFish server:

```
$GF_HOME/bin/asadmin stop-domain
%GF_HOME%\bin\asadmin stop-domain # Windows
```

### External context

In time, it has been proven that it's useful to keep the contents of the war archive on a location on the disk, and point the application to the folder fith the contents. This is called exploded (exploded war) deployment. 

To perform this type of deployment, first build the application from the `homework` directory:

```
gradle clean build
```

Next, since this type of deployment requires unzipping the war archive, extract the contents of the war archive inside `extern/exploded_war`.

For Tomcat, copy the `extern/aliasTomcat.xml` file to a special location in the tomcat directory:

```
cp extern/aliasTomcat.xml $CATALINA_HOME/conf/Catalina/localhost/homework.xml
copy extern\aliasTomcat.xml %CATALINA_HOME%\conf\Catalina\localhost\homework.xml # Windows
```

Run the Tomcat container using

```
$CATALINA_HOME/bin/startup.sh
%CATALINA_HOME%\bin\startup.bat # Windows
```

Open the browser and access the application at `http://localhost:8080/homework/home`.

In order to stop the Tomcat container, use the command:

```
$CATALINA_HOME/bin/shutdown.sh
%CATALINA_HOME%\bin\shutdown.bat # Windows
```

Finally, to perform a cleanup operation just remove the xml configuration file that was uploaded inside Tomcat's directory. In order to do this, run 

```
rm  $CATALINA_HOME/conf/Catalina/localhost/homework.xml 
del  %CATALINA_HOME%\conf\Catalina\localhost\homework.xml # Windows
```

For Jetty, the process is similar. Copy the aliasJetty.xml file to the Jetty home directory (the file can be named anything with the .xml extension):

```
cp extern/aliasJetty.xml $JETTY_HOME/webapps/homework.xml
copy extern\aliasJetty.xml %JETTY_HOME%\webapps\homework.xml # Windows
```

Next, start the Jetty container. Run this command from the `$JETTY_HOME` directory (TODO: for some reason, doesn't work outside):

```
java -jar start.jar
java -jar start.jar
```

Open the browser and access the application at `http://localhost:8080/homework/home`. To stop the container, just kill its running process in the terminal.

Finally, to perform a cleanup operation just remove the xml configuration file that was uploaded inside Jetty's directory. In order to do this, run 

```
rm extern/aliasJetty.xml $JETTY_HOME/webapps/homework.xml
del extern\aliasJetty.xml %JETTY_HOME%\webapps\homework.xml # Windows
```

### Embedded deployments

In case of embedded deployments, the server configuration is no longer creader as an xml file, but rather integrated in the code directly. The projects now have two classes `TomcatServer.java` and `JettyServer.java` which configure the Tomcat and Jetty server respectively. The `build.gradle` file also changed to include the embedded dependencies and apply the java plugin instead (this type of deployment produces an executable jar archive, not a web archive).

To get started, build the application with

```
gradle clean build
```

Now, the server (in both Tomcat and Jetty cases) is ran by just executing the jar archive:

```
java -jar build/libs/homeworkembedded.jar
java -jar build\libs\homeworkembedded.jar # Windows
```

The server can be stopped by killing its running process in the terminal.

TODO: Jetty doesn't work yet.

## Errors encountered

1. Jetty wouldn't want to start the app. The error encountered was 

```
2022-11-27 12:32:59.712:INFO:oejsh.ContextHandler:Scanner-0: Stopped o.e.j.w.WebAppContext@398ea2c6{homework,/homework,null,STOPPED}{C:\jetty-distribution-9.4.49.v20220914\webapps\homework.war}
2022-11-27 12:32:59.743:WARN:oejw.WebAppContext:Scanner-0: Failed startup of context o.e.j.w.WebAppContext@5fbbb19b{homework,/homework,file:///C:/Users/alexb/AppData/Local/Temp/jetty-0_0_0_0-8080-homework_war-_homework-any-8136321115002338320/webapp/,STOPPED}{C:\jetty-distribution-9.4.49.v20220914\webapps\homework.war}
java.lang.NumberFormatException: For input string: "1.0"
        at java.lang.NumberFormatException.forInputString(Unknown Source)
        at java.lang.Integer.parseInt(Unknown Source)
        at java.lang.Integer.parseInt(Unknown Source)
        at org.eclipse.jetty.webapp.WebDescriptor.processVersion(WebDescriptor.java:253)
        at org.eclipse.jetty.webapp.WebDescriptor.parse(WebDescriptor.java:214)
        at org.eclipse.jetty.webapp.MetaData.setWebXml(MetaData.java:193)
        at org.eclipse.jetty.webapp.WebXmlConfiguration.preConfigure(WebXmlConfiguration.java:55)
        at org.eclipse.jetty.webapp.WebAppContext.preConfigure(WebAppContext.java:488)
        at org.eclipse.jetty.webapp.WebAppContext.doStart(WebAppContext.java:523)
        at org.eclipse.jetty.util.component.AbstractLifeCycle.start(AbstractLifeCycle.java:73)
        at org.eclipse.jetty.deploy.bindings.StandardStarter.processBinding(StandardStarter.java:46)
        at org.eclipse.jetty.deploy.AppLifeCycle.runBindings(AppLifeCycle.java:188)
        at org.eclipse.jetty.deploy.DeploymentManager.requestAppGoal(DeploymentManager.java:517)
        at org.eclipse.jetty.deploy.DeploymentManager.addApp(DeploymentManager.java:157)
        at org.eclipse.jetty.deploy.providers.ScanningAppProvider.fileChanged(ScanningAppProvider.java:190)
        at org.eclipse.jetty.deploy.providers.WebAppProvider.fileChanged(WebAppProvider.java:401)
        at org.eclipse.jetty.deploy.providers.ScanningAppProvider$1.fileChanged(ScanningAppProvider.java:72)
        at org.eclipse.jetty.util.Scanner.reportChange(Scanner.java:827)
        at org.eclipse.jetty.util.Scanner.reportDifferences(Scanner.java:757)
        at org.eclipse.jetty.util.Scanner.scan(Scanner.java:641)
        at org.eclipse.jetty.util.Scanner$1.run(Scanner.java:558)
        at java.util.TimerThread.mainLoop(Unknown Source)
        at java.util.TimerThread.run(Unknown Source)
```

This had been solved by updating the `web-app` properties in the xml file from

```xml
<web-app xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xmlns="http://java.sun.com/xml/ns/javaee"
         xsi:schemaLocation="http://java.sun.com/xml/ns/javaee
         http://java.sun.com/xml/ns/javaee/web-app_3_1_0.xsd"
         version="3.1.0">
```

to 

```xml
<web-app xmlns="http://java.sun.com/xml/ns/javaee" 
         xmlns:xsi="http://www.w3.org/2001/XMLSchema-instance"
         xsi:schemaLocation="http://java.sun.com/xml/ns/javaee
         http://java.sun.com/xml/ns/javaee/web-app_3_0.xsd"
         version="3.0">
```

2. Glassfish for some reason doesn't want to see an empty url-pattern for the dispatch servlet. I changed it to home and it seems to work, but not sure nor care why that is happening.
