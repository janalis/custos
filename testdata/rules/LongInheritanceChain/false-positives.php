<?php
namespace App {
    class Node {}
    class Branch extends Node {}
    class Twig extends Branch {}

    /** @deprecated */
    class OldLeaf extends Twig {}
    class TwigException extends Twig {}

    abstract class Base extends Branch {}
    class Concrete extends Base {}

    class Err1 extends \RuntimeException {}
    class Err2Exception extends Err1 {}
    class Err3 extends Err2Exception {}
    class Err4 extends Err3 {}

    class Loose extends Missing {}
    class Shallow extends Branch {}
}
namespace App\Tests {
    class Node {}
    class Branch extends Node {}
    class Twig extends Branch {}
    class Leaf extends Twig {}
}
