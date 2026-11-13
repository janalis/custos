<?php
namespace App;

// Composer installs dependencies into libs/: a library's hierarchy is its own.
class Widget extends \Acme\Kit\Base {}
// legacy/vendor/ is a directory of the project, not the vendor directory.
class <weak_warning descr="3 levels of parent classes; prefer composition over deep inheritance.">Post</weak_warning> extends \Legacy\Model {}
