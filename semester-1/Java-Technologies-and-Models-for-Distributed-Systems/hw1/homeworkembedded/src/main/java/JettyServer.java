import org.eclipse.jetty.server.Server;
import org.eclipse.jetty.servlet.ServletHandler;

public class JettyServer {
    public static void main(String[] args) throws Exception {
        Server server = new Server(8080);
        ServletHandler handler = new ServletHandler();
        server.setHandler(handler);

        handler.addServletWithMapping(homework.Dispatcher.class, "/home");
        handler.addServletWithMapping(homework.ClientHtml.class, "/client");
        handler.addServletWithMapping(homework.ErrorHtml.class, "/error");
        handler.addServletWithMapping(homework.StoreNumber.class, "/storeNumber");
        System.out.println("Start server Jetty embedded");
        server.start();
        server.join();
    }
}
