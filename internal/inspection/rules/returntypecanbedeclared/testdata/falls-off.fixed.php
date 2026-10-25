<?php
class Picker {
    public function partial($f) {
        if ($f) {
            return new \ArrayObject();
        } elseif ($f === 0) {
            return null;
        }
    }

    public function picked($k) {
        switch ($k) {
            case 1:
                return new \ArrayObject();
        }
    }

    public function either($f): ?\ArrayObject {
        if ($f) {
            return new \ArrayObject();
        } else {
            return null;
        }
    }

    public function idle($f): void {
        if ($f) {
            $this->n = 1;
        }
    }
}
