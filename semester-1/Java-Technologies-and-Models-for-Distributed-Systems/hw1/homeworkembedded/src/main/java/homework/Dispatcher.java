package homework;

import javax.servlet.*;
import javax.servlet.http.*;
import java.io.*;

public class Dispatcher extends HttpServlet {
    public void doGet(HttpServletRequest req, HttpServletResponse res)
            throws ServletException, IOException {

        System.out.println("sdsadsadasdas");
        RequestDispatcher rdStoreNumber = req.getRequestDispatcher("storeNumber");
        System.out.println("fffffffffff");
        RequestDispatcher rdError = req.getRequestDispatcher("error");
        RequestDispatcher rdIndex = req.getRequestDispatcher("client");

        // If any parameter is null, show the index
        if (req.getParameter("name") == null || req.getParameter("number") == null) {
            rdIndex.forward(req, res);

            // If any of these is invalid, get lost
        } else if (req.getParameter("name") == "") {
            rdError.forward(req, res);
        } else {
            rdStoreNumber.forward(req, res);
        }

    }

    public void doPost(HttpServletRequest req, HttpServletResponse res)
            throws ServletException, IOException {
        doGet(req, res);
    }
}
