package homework;

import java.io.*;
import java.util.HashMap;
import java.util.Map;

import javax.servlet.*;
import javax.servlet.http.*;
import javax.servlet.annotation.*;

class PropertyCounter {
    private boolean isStartingWithSix(int x) {
        while (x > 10) {
            x = x / 10;
        }
        return x == 6;
    }

    private boolean isEndingWithNine(int x) {
        return x % 10 == 9;
    }

    private boolean hasTwoDigits(int x) {
        return 9 < x && x < 100;
    }

    public int Count(int x) {
        int count = 0;
        if (this.isEndingWithNine(x)) {
            count++;
        }
        if (this.isStartingWithSix(x)) {
            count++;
        }
        if (this.hasTwoDigits(x)) {
            count++;
        }
        return count;
    }

}

class Sorter {
    public String[] getSortedNames(Map<String, Integer> scores) {
        String[] names = scores.keySet().toArray(new String[0]);

        // Sort by ascending score
        for (int i = 0; i < names.length - 1; i++) {
            for (int j = i + 1; j < names.length; j++) {
                if (scores.get(names[i]) < scores.get(names[j])) {
                    String aux = names[i];
                    names[i] = names[j];
                    names[j] = aux;
                }
            }
        }

        // Also sort names
        for (int i = 0; i < names.length - 1; i++) {
            for (int j = i + 1; j < names.length; j++) {
                // If they have equal score, swap names
                if (scores.get(names[i]) == scores.get(names[j])) {
                    if (names[i].compareTo(names[j]) > 0) {
                        String aux = names[i];
                        names[i] = names[j];
                        names[j] = aux;
                    }
                }
            }
        }

        return names;
    }
}

public class StoreNumber extends HttpServlet {

    private Map<String, Integer> numberMap = new HashMap<>();

    // Dependency injection via field injection
    private PropertyCounter pc = new PropertyCounter();
    private Sorter sorter = new Sorter();

    protected void doGet(HttpServletRequest request,
            HttpServletResponse response) throws ServletException, IOException {

        String name;
        int number;

        // Name has already been validated
        name = request.getParameter("name");

        // Number has already been validated
        number = Integer.parseInt(request.getParameter("number"));

        int count = pc.Count(number);

        // Find the previous score of that person
        Integer previousScore = numberMap.get(name);
        if (previousScore == null) {
            numberMap.put(name, count);
        } else {
            if (count > previousScore) {
                // Update the score if it's better.
                numberMap.put(name, count);
            }
        }

        String[] leaderboard = this.sorter.getSortedNames(numberMap);

        response.setContentType("text/html");
        PrintWriter out = response.getWriter();
        out.println("<!DOCTYPE html>");
        out.println("<html>");
        out.println("<head>");
        out.println("<title>Guess the property</title>");
        out.println("</head>");
        out.println("<body>");
        out.println("<h2>Leaderboard</h2>");
        out.println("<ul>");
        for (String currentName : leaderboard) {
            out.println("<li>");
            out.println(currentName + ": " + numberMap.get(currentName).toString());
            out.println("</li>");
        }
        out.println("</ul>");
        out.println("</body>");
        out.println("</html>");
        out.close();
    }

    protected void doPost(HttpServletRequest request,
            HttpServletResponse response) throws ServletException, IOException {
        doGet(request, response);
    }
}
