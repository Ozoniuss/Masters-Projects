package homework;

import javax.servlet.*;
import javax.servlet.http.*;
import java.io.*;

public class Dispatcher extends HttpServlet {
    public void doGet(HttpServletRequest req, HttpServletResponse res)
            throws ServletException, IOException {

        RequestDispatcher rdStoreNumber = req.getRequestDispatcher("storeNumber");
        RequestDispatcher rdError = req.getRequestDispatcher("error");

        RequestDispatcher rdIndex = req.getRequestDispatcher("client");

        // If both parameters are null, show the index
        if (req.getParameter("name") == null && req.getParameter("number") == null) {
            rdIndex.forward(req, res);

            // If name is not provided
        } else if (req.getParameter("name").isEmpty() || req.getParameter("name") == null) {
            rdError.forward(req, res);
        } else {

            try {
                int number = Integer.parseInt(req.getParameter("number"));
                rdStoreNumber.forward(req, res);
            } catch (Exception e) {
                rdError.forward(req, res);
            }
        }

    }

    public void doPost(HttpServletRequest req, HttpServletResponse res)
            throws ServletException, IOException {
        doGet(req, res);
    }
}
