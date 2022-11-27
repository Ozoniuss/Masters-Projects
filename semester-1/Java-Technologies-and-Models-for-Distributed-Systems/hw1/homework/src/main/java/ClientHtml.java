package homework;

import javax.servlet.*;
import javax.servlet.http.*;
import java.io.*;

public class ClientHtml extends HttpServlet {

    protected void doGet(HttpServletRequest request,
            HttpServletResponse response) throws IOException {

        response.setContentType("text/html");
        PrintWriter out = response.getWriter();

        out.print("<!DOCTYPE html><html>\n" +
                "<head><title>Suma sau conversie</title></head>\n" +
                "<body>\n" +
                "<form method=\"GET\"action=\"\">\n" +
                "Name:<input type=\"text\"name=\"name\">\n" +
                "Number:<input type = \"text\"name=\"number\">\n" +
                "<input type = \"submit\"value=\"Pls add number.\"/>\n" +
                "</form>\n" +
                "</body>\n" +
                "</html>\n");
        out.close();

    }
}
