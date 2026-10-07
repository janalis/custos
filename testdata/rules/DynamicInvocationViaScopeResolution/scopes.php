<?php

interface Shape
{
    public function area();
}

class Square implements Shape
{
    public function area() {}
    public function side() {}
}

class Unrelated
{
    public function side() {}

    public function measure(Shape $shape, $unknown)
    {
        Square::area();
        Square::side();
        Shape::area();
        $unknown::side();
        Square::$method();
    }

    public function other()
    {
        Square::area();
    }
}

function recovery()
{
    Square::();
}
