# Running java

First, let's write some code in the `SomeClass` java file. Note that this class must define exactly one public class named `SomeClass`, exactly like the file name. In order to run the code, it:

1. Must be compiled to byte code (`.class`)
2. Byte code can then be executed on any operating system with a JRE (Java Runtime Environment) installed.

## Compilation

Compilation can be done with the following command:

```
javac SomeClass.java
```

If other classes (not public ones) are defined inside the `.java` file, a `.class` file with their compiled code is also generated.

## Execution

The compiled by code is translated to the operating system's native code via the JVM found inside the installed JRE. 

In order to be able to execute the code, a `main` function must be define inside the java class. Then, the code can be executed with the command:

```
java SomeClass
```
