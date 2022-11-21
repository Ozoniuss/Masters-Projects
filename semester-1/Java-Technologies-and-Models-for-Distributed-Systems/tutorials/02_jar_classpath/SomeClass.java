import org.apache.log4j.*;

public class SomeClass {
    static Logger log = Logger.getLogger(SomeClass.class);

    public static void main(String[] args) {
        log.info("1");
        System.out.println("Hello World!");
        log.info("2");
    }
}
