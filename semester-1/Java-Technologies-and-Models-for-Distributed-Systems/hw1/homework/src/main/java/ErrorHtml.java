package homework;

import javax.servlet.*;
import javax.servlet.http.*;
import java.io.*;

public class ErrorHtml extends HttpServlet {

    protected void doGet(HttpServletRequest request,
            HttpServletResponse response) throws IOException {

        response.setContentType("text/html");
        PrintWriter out = response.getWriter();
        out.println("<!DOCTYPE html>");
        out.println("<html>");
        out.println("<head>");
        out.println("<title>Error</title>");
        out.println("</head>");
        out.println("<body>");
        out.println("<h2>You had an error, fix it</h2>");
        if (request.getParameter("name") == null || request.getParameter("name").isEmpty()) {
            out.println("<h3>don't u have a name??<h3>");
        } else {
            out.println("<h3>You know that \"" + request.getParameter("number") + "\" is not a number right? </h3>");
        }
        out.println("</body>");
        out.println("</html>");
        out.close();
    }

}
