<?php
class A {}
class B extends A {}
class <weak_warning descr="2 levels of parent classes; prefer composition over deep inheritance.">C</weak_warning> extends B {}
class D extends A {}
