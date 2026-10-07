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
        $this->paint("red", 2);
        self::registry();
        parent::resize();
        parent::missing();
        $this->paint('y');
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
        $this->resize();
    }
}
