<?php
class HookSplice {
    public mixed $p {
        get {
            $items = [];
            unrelated();
            <warning descr="Preserve replacement keys before reading the inserted key.">array_splice($items, 0, 0, ['name' => 'Ada'])</warning>;
            echo $items['name'];
        }
    }
}
