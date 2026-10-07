<?php
interface Shape
{
    public function scale(<weak_warning descr="Prefer null as the default value for this parameter.">$factor = 1</weak_warning>);
}

abstract class Figure implements Shape
{
}

class Circle extends Figure
{
    public function scale(<weak_warning descr="Prefer null as the default value for this parameter.">$factor = 1</weak_warning>) {}
}

function dnf(<weak_warning descr="Prefer null as the default value for this parameter.">(Countable&Traversable)|null $items = []</weak_warning>, (Countable&Traversable)|array $more = []) {}
