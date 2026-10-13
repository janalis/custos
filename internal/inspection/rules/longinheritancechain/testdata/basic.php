<?php
namespace App {
    class Node {}
    class Branch extends Node {}
    class Twig extends Branch {}
    class <weak_warning descr="3 levels of parent classes; prefer composition over deep inheritance.">Leaf</weak_warning> extends Twig {}
    class <weak_warning descr="4 levels of parent classes; prefer composition over deep inheritance.">Bud</weak_warning> extends Leaf {}
    abstract class <weak_warning descr="3 levels of parent classes; prefer composition over deep inheritance.">AbstractLeaf</weak_warning> extends Twig {}
}

namespace yii\base {
    class Component {}
}
namespace App\Widgets {
    class Panel extends \yii\base\Component {}
    class FancyPanel extends Panel {}
    class <weak_warning descr="3 levels of parent classes; prefer composition over deep inheritance.">MegaPanel</weak_warning> extends FancyPanel {}
}
