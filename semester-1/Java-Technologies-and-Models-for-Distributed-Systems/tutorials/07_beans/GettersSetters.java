/* A class representing a java bean */

public class GettersSetters {
    private String name;

    // Simple methods
    public String getName() {
        return name;
    }

    public void setName(String name) {
        this.name = name;
    }

    // Boolean Methods (are tied to boolean fields)
    private boolean old;

    public boolean isOld() {
        return old;
    }

    // Index methods
    private int[] numbers;

    public int[] getNumbers() {
        return numbers;
    }

    public int getNumbern(int i) {
        return numbers[i];
    }

    public void setNumbers(int[] numbers) {
        this.numbers = numbers;
    }

    public void setNumbers(int i, int number) {
        this.numbers[i] = number;
    }

    public static void main(String[] args) {
        GettersSetters s = new GettersSetters();
        s.setName("one");
        System.out.printf("name: %s\n", s.getName());
        System.out.printf("isold: %b\n", s.isOld());
    }
}