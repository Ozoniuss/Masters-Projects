import java.io.*;
import org.apache.catalina.*;
import org.apache.catalina.startup.*;

import homework.ClientHtml;
import homework.Dispatcher;
import homework.ErrorHtml;
import homework.StoreNumber;

public class TomcatServer {
    public static void main(String[] args) throws Exception {
        Tomcat server = new Tomcat();
        server.setPort(8080);
        Context ctx = server.addContext("", (new File(".")).getAbsolutePath());
        Tomcat.addServlet(ctx, "client", new ClientHtml());
        Tomcat.addServlet(ctx, "storeNumber", new StoreNumber());
        Tomcat.addServlet(ctx, "error", new ErrorHtml());
        Tomcat.addServlet(ctx, "dispatcher", new Dispatcher());

        ctx.addServletMappingDecoded("/home", "dispatcher");
        ctx.addServletMappingDecoded("/client", "client");
        ctx.addServletMappingDecoded("/error", "error");
        ctx.addServletMappingDecoded("/storeNumber", "storeNumber");
        server.start();
        System.out.println("Start server Tomcat embedded");
        server.getServer().await();
    }
}
