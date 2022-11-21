# Fatjar solution for having all dependencies in a single jar file

The idea is pretty straightforward: all jar dependencies are extracted and then a single jar file is created.

```
# Compile the java file with dependency in classpath 
javac -cp log4j-1.2.17.jar SomeClass.java
# Unzip log4j jar dependency
unzip log4j-1.2.17.jar
# Create the manifest with the main file
echo Main-Class: SomeClass >> manifest.mf
# Create the executable jar archive which includes dependencies
jar -cmf manifest.mf new.jar SomeClass.class org/ log4j.properties
# Remove files from log4j jar archive and cleanup
rm -rf org
rm -rf META-INF
rm manifest.mf
rm SomeClass.class

# Execute the new jar archive
java -jar new.jar
```

A more clever approach would be to include just the classes used by the application (by exploring the imports).

Note that it is possible this approach leads to name conflicts.