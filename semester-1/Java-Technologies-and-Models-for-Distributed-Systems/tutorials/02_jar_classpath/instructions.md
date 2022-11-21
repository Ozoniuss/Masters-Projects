# Java Archives

## Archives

Java applications as well as java-based technologies and application servers can be packed into archives. The archive types are:

- jar (Java Archive)
- war (Web Archive)
- ear (Enterprise Archive)
- rar (Resource Archive)

These are all .zip archives packing a specific directory like structure. Each archive comes with specific tools for creating and managing it.

## Classpath

The `CLASSPATH` environment variable contains a list of directories and jar archives that are explored by the JVM (ClassLoader). Its value is set by the Java environment, but can also be overwritten like any other environment variable. Its content can also be overwritten by specifying the `-cp` flag: `java -cp`. Alternatively, the `Class-Path` attribute can be set inside `META-INF/MANIFEST.MF` inside a jar archive.

Note that `Class-Path` inside a manifest refers to jar archives found on the filesystem, not inside jar archives! For this reason, the ClassLoader doesn't see the content of jars inside jars.

Note that with the `-cp` flag, classpath can be specified with wildcards, but that doesn't work if used in the manifest file.

1. Create a new jar archive with:

```
jar cf 01new.jar Simple.java 
```

2. This creates a jar archive with the log4j-1.2.17.jar dependency. The class `SomeClass.java` requiures the log4j dependency and it cannot be compiled directly:

<details>
<summary><b>Compilation error</b></summary>

```
$ javac SomeClass.java 
SomeClass.java:1: error: package org.apache.log4j does not exist
import org.apache.log4j.*;
^
SomeClass.java:4: error: cannot find symbol
    static Logger log = Logger.getLogger(SomeClass.class);
           ^
  symbol:   class Logger
  location: class SomeClass
SomeClass.java:4: error: cannot find symbol
    static Logger log = Logger.getLogger(SomeClass.class);
                        ^
  symbol:   variable Logger
  location: class SomeClass
3 errors
```

</details>

This is because the log4j jar dependency must be specified. One possible way is to specify the `classpath` during compilation:

```
javac -cp '.;log4j-1.2.17.jar' SomeClass.java
```

In order to execute this file, the same classpath must be specified again:

```
java -cp '.;log4j-1.2.17.jar' SomeClass
```

3. The following specifies `Class-Path` and `Main-Class` in a text file and provides that text file for the manifest. The classpath includes the log4j jar dependency, and a jar file for the SomeClass which requires that dependency can be created as:

```
# compile the java code
javac -cp log4j-1.2.17.jar SomeClass.java
# create the jar file with the manifest containing main-class and class-path
jar cmf 03manifest.txt 03new.jar SomeClass.class log4j.properties
# run the jar file
java -jar 03new.jar
```

Note that moving the log4j jar dependency to another location outside this directory will cause the code to break. The same error is thrown if the dependency location is not found in `Class-Path` the .jar archive (make sure to have an endline):

<details>
<summary><b>NoClassDefFoundError</b></summary>

```
$ java -jar 03new.jar 
Exception in thread "main" java.lang.NoClassDefFoundError: org/apache/log4j/Logger
        at SomeClass.<clinit>(SomeClass.java:4)
Caused by: java.lang.ClassNotFoundException: org.apache.log4j.Logger
        at java.net.URLClassLoader.findClass(Unknown Source)
        at java.lang.ClassLoader.loadClass(Unknown Source)
        at sun.misc.Launcher$AppClassLoader.loadClass(Unknown Source)
        at java.lang.ClassLoader.loadClass(Unknown Source)
        ... 1 more
```

</details>

Removing the `Main-Class` attribute also causes the execution to fail.

### Caution

It might be reasonable to think that if `Class-Path` inside manifest doesn't include the log4j dependency, a command like 

```
java -cp <log4j-jar-path>.jar -jar 03new.jar 
```

but this doesn't in fact work. THat's because `-jar` and `-cp` exclude each other.

Similarly, the log4j jar dependency cannot be copied inside the resulting .jar archive, because the ClassLoader doesn't look inside jar archives. This is an inconenience, because jar dependencies must always exist on the local system (under that specific path).