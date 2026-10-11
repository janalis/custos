<?php

// GMP converts to string through an internal handler, without a __toString method.
array_unique([new GMP(5), new GMP(5), new GMP(7)]);
array_unique([new \GMP('12'), new \GMP('12')], SORT_STRING);

class Labelled
{
    public function __toString()
    {
        return 'label';
    }
}

array_unique([new Labelled(), new Labelled()]);

// The hierarchy is not fully known.
class Orphan extends NotDeclaredAnywhere
{
}

array_unique([new Orphan(), new Orphan()]);

class Blended
{
    use NotDeclaredTrait;
}

array_unique([new Blended(), new Blended()]);
