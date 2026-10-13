<?php
class Person {} $p=new Person(); <warning descr="Declare this property before assigning it.">$p->label</warning>='Mia';
