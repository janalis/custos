<?php
class SignatureBase {
    public function <weak_warning descr="Declare ': int' as the return type (update the whole hierarchy with a signature refactoring).">count</weak_warning>() { return 1; }
}
class SignatureChild extends SignatureBase {
    use UnavailableMethods;
}
