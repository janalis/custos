<?php
class Statement
{
    public function __destruct()
    {
        echo 'closed';
    }
}

class Prepared extends Statement {}

class Plain {}

final class Result
{
    private ?Statement $statementReference;
    private ?Prepared $prepared;
    private Plain $plain;

    public function __construct(?Statement $statementReference, Prepared $prepared, Plain $plain)
    {
        // Holding the statement delays its destructor.
        $this->statementReference = $statementReference;
        $this->prepared = $prepared;
        <weak_warning descr="Private property is only used in the constructor; likely dead code.">$this->plain</weak_warning> = $plain;
    }
}
