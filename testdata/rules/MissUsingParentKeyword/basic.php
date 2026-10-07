<?php

class Widget
{
    public function paint($color) {}
    public static function registry() {}
    public function resize() {}
    public function layout() {}
}

class Button extends Widget
{
    public function layout()
    {
        parent::layout();
        <weak_warning descr="Call it as '$this->paint(&quot;red&quot;, 2)' instead of through 'parent::'.">parent::paint("red", 2)</weak_warning>;
        <weak_warning descr="Call it as 'self::registry()' instead of through 'parent::'.">parent::registry()</weak_warning>;
        parent::resize();
        parent::missing();
        <weak_warning descr="Call it as '$this->paint(&apos;y&apos;)' instead of through 'parent::'.">Parent::paint('y')</weak_warning>;
        PARENT::Layout();
        $fn = function () { parent::paint('x'); };
    }

    public static function build()
    {
        parent::paint('blue');
    }
}

class IconButton extends Button
{
    public function resize() {}
}

final class Toggle extends Widget
{
    public function layout()
    {
        <weak_warning descr="Call it as '$this->resize()' instead of through 'parent::'.">parent::resize()</weak_warning>;
    }
}
