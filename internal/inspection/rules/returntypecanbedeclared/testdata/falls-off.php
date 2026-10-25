<?php
class Picker {
    public function <weak_warning descr="Declare ': ?\ArrayObject' as the return type.">partial</weak_warning>($f) {
        if ($f) {
            return new \ArrayObject();
        } elseif ($f === 0) {
            return null;
        }
    }

    public function <weak_warning descr="Declare ': ?\ArrayObject' as the return type.">picked</weak_warning>($k) {
        switch ($k) {
            case 1:
                return new \ArrayObject();
        }
    }

    public function <weak_warning descr="Declare ': ?\ArrayObject' as the return type.">either</weak_warning>($f) {
        if ($f) {
            return new \ArrayObject();
        } else {
            return null;
        }
    }

    public function <weak_warning descr="Declare ': void' as the return type.">idle</weak_warning>($f) {
        if ($f) {
            $this->n = 1;
        }
    }
}
