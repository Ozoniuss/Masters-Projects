import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.servlet.ServletContextHandler;
import org.eclipse.jetty.servlet.ServletHandler;

// public class JettyServer {
//     public static void main(String[] args) throws Exception {
//         Server server = new Server(8080);
//         ServletHandler handler = new ServletHandler();
//         server.setHandler(handler);

//         handler.addServletWithMapping(homework.Dispatcher.class, "/home");
//         handler.addServletWithMapping(homework.ClientHtml.class, "/client");
//         handler.addServletWithMapping(homework.ErrorHtml.class, "/error");
//         handler.addServletWithMapping(homework.StoreNumber.class, "/storeNumber");
//         System.out.println("Start server Jetty embedded");
//         server.start();
//         server.join();
//     }
// }

public class JettyServer {
    public static void main(String[] args) throws Exception {
        Server server = new Server(8080);
        // ServletHandler handler = new ServletHandler();
        // server.setHandler(handler);

        ServletContextHandler context = new ServletContextHandler();
        context.setContextPath("/");
        server.setHandler(context);

        // context.addServlet("Dispatcher", "");
        context.addServlet("homework.Dispatcher", "/home");
        context.addServlet("homework.ClientHtml", "/client");
        context.addServlet("homework.ErrorHtml", "/error");
        context.addServlet("homework.StoreNumber", "/storeNumber");
        System.out.println("Start server Jetty embedded");
        server.start();
        server.join();
    }
}