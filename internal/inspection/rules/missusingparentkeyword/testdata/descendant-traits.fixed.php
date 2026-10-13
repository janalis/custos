<?php

class Canvas
{
    public function paint() { return 'Canvas::paint'; }
    public function size() { return 'Canvas::size'; }
    public function color() { return 'Canvas::color'; }
    public function name() { return 'Canvas::name'; }
}

class Panel extends Canvas
{
    public function draw()
    {
        // Sidebar imports paint() from a trait, Footer aliases a trait method to size().
        return [
            parent::paint(),
            parent::size(),
            parent::color(),
            $this->name(),
        ];
    }
}

trait Paints
{
    public function paint() { return 'Paints::paint'; }
}

trait Tints
{
    use Shades;
}

trait Shades
{
    public function color() { return 'Shades::color'; }
}

trait Measures
{
    public function measure() { return 'Measures::measure'; }
}

class Sidebar extends Panel
{
    use Paints;
}

class Header extends Panel
{
    use Tints;
}

class Footer extends Panel
{
    use Measures {
        measure as size;
    }
}

echo json_encode([(new Sidebar)->draw(), (new Header)->draw(), (new Footer)->draw()]), "\n";
