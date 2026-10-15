<?php
class HookLabels {
    public mixed $p {
        get {
            goto used;
            used:
            <weak_warning descr="Label 'unused' is never targeted by a goto; remove it.">unused:</weak_warning>
            return 1;
        }
    }
}
