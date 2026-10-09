<?php
class HookedInitialization
{
    private int $setter = 0 {
        set { echo $value; $this->setter = $value; }
    }
    private int $getter = 0 {
        get { echo 'read'; return $this->getter; }
    }
    private int $both = 0 {
        get => $this->both;
        set => $value;
    }
    private $ordinary;
    private int $ordinaryTyped = 0;

    public function __construct()
    {
        $this->setter = 0;
        $this->getter = 0;
        $this->both = 0;
        $this->setter = 2;
        $this->getter = 2;
        $this->both = 2;
        $this->ordinary = 2;
        $this->ordinaryTyped = 0;
    }
}
