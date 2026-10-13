<?php
function counter() {
    static $n = 0;
    return $n = $n + 1;
}

function guarded() {
    try {
        return $res = compute();
    } finally {
        release($res);
    }
}

function wrapped() {
    return ($tmp = compute());
}

function compound($v) {
    return $v .= 'x';
}

function glob() {
    global $cfg;
    return $cfg = load();
}

function shared() {
    $later = function () use (&$val) { return $val; };
    return $val = 1;
}

function props($o) {
    if ($o) {
        return;
    }
    return $o->p = 1;
}

class Grid {
    public $rows = [1, 2];
    public $last = 0;

    function first() {
        foreach ($this->rows as &$row) {
            return $row = 9;
        }
    }

    function keyed() {
        foreach ($this->rows as $i => &$cell) {
            return $cell = 9;
        }
    }

    function mark() {
        $slot = &$this->last;
        return $slot = 5;
    }

    function aliased() {
        $this->alias = &$mine;
        return $mine = 5;
    }

    function unpacked() {
        [&$head] = $this->rows;
        return $head = 5;
    }

    function listed() {
        list('k' => &$item) = $this->rows;
        return $item = 5;
    }

    function collected() {
        $this->refs = [&$kept];
        return $kept = 5;
    }
}

function slot(array &$store) {
    return $cell = &$store['k'];
}

function outerFinally() {
    try {
        try {
            return $res = 7;
        } catch (\LogicException $e) {
        }
    } finally {
        echo $res ?? 'nothing';
    }
}

return $top = 1;
