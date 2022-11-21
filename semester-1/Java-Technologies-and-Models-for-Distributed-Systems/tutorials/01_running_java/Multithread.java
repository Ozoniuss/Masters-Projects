// Java code for thread creation by extending
// the Thread class
class MultithreadingDemo extends Thread {

    public MultithreadingDemo(String name) {
        super(name);
    }

    public void run() {
        try {
            // Displaying the thread that is running
            System.out.println(
                    "Thread " + Thread.currentThread().getName()
                            + " is running");

            sleep(1000);

            // Displaying the thread that is running
            System.out.println(
                    "Thread " + Thread.currentThread().getName()
                            + " has stopped sleeping");

        } catch (Exception e) {
            // Throwing an exception
            System.out.println("Exception is caught");
        }
    }
}

// Main Class
public class Multithread {
    public static void main(String[] args) {
        int n = 8; // Number of threads
        for (int i = 0; i < n; i++) {
            MultithreadingDemo object = new MultithreadingDemo(String.format("t%s", i));

            // start invokes the run method of the Thread object
            object.start();

        }
        try {
            Thread.sleep(3000);
            System.out.println("Main thread has stopped sleeping");
        } catch (Exception e) {
            // Throwing an exception
            System.out.println("Exception is caught in main");
        }
    }
}