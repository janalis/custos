<?php
class ReturnedHookLocal {
    public mixed $reference {
        &get {
            $original = 5;
            return $local = $original;
        }
    }
    public int $value {
        get {
            return 5;
        }
    }
}
