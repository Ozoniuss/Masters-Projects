public class Control {
    public static void main(String[] args) {

        // Direct control
        BaseDirect bd = new BaseDirect();
        bd.getMember().setValue("base direct");
        bd.getMember().printMe();

        // DI
        BaseDISetter bdis = new BaseDISetter();
        bdis.setMember(new Member());
        bdis.getMember().setValue("base dependency injection setter");
        bdis.getMember().printMe();

        BaseDIConstructor bdic = new BaseDIConstructor(new Member());
        bdic.getMember().setValue("base dependency injection constructor");
        bdic.getMember().printMe();

        BaseDIProperty bdip = new BaseDIProperty();
        bdip.member = new Member();
        bdip.getMember().setValue("base dependency injection property");
        bdip.getMember().printMe();
    }
}

// Base bean, direct control
class BaseDirect {
    private Member member = null;

    public BaseDirect() {
        this.member = new Member();
    }

    public Member getMember() {
        return this.member;
    }
}

// Base bean, Dependency Injection via setter;
class BaseDISetter {
    private Member member = null;

    public Member getMember() {
        return this.member;
    }

    public void setMember(Member member) {
        this.member = member;
    }
}

// Base bean, Dependency Injection via constructor;
class BaseDIConstructor {
    private Member member = null;

    public BaseDIConstructor(Member member) {
        this.member = member;
    }

    public Member getMember() {
        return this.member;
    }

}

// Base bean, Dependency Injection via property;
class BaseDIProperty {
    public Member member = null;

    public Member getMember() {
        return this.member;
    }
}

// Member bean

class Member {
    private String value;

    public String getValue() {
        return this.value;
    }

    public void setValue(String value) {
        this.value = value;
    }

    public void printMe() {
        System.out.printf("Hi I'm a member from class %s.\n", this.value);
    }
}
