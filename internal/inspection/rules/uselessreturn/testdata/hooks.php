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
            <weak_warning descr="The assigned variable is never used after returning; return the value directly.">return $local = 5;</weak_warning>
        }
    }
}
