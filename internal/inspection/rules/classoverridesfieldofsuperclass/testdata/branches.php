<?php
class Shape
{
    protected int $sides;
}

class Square extends Shape
{
    protected int <weak_warning descr="Property 'sides' is already declared in \Shape; drop this re-declaration.">$sides</weak_warning>;
}

class Broken extends Shape
{
    // Recovered declaration without a variable name.
    public int ;
}
