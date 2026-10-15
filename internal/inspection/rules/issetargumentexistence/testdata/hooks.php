<?php
class HookChecks {
    public mixed $p {
        get {
            return isset(<error descr="Variable '$missing' is not defined in this scope.">$missing</error>);
        }
        set {
            if (isset($value)) { consume($value); }
        }
    }
}
