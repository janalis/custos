<?php

class Meter
{
    private static int $writes = 0;
    public static int $limit = 10;

    public int $value = 0 {
        set {
            static::$writes = static::$writes + 1;
            Meter::$writes += 0;
            <weak_warning descr="Modify this static property only from the class that declares it.">Gauge::$unit = 'kg'</weak_warning>;
            Missing::$thing = 1;
            $bump = fn () => Meter::$limit = 20;
            $bump();
            $this->value = $value;
        }
    }

    public function __construct(
        public string $label = '' {
            set {
                static::$writes = 0;
                <weak_warning descr="Modify this static property only from the class that declares it.">Gauge::$unit = $value</weak_warning>;
                $this->label = $value;
            }
        },
    ) {
    }
}

class Gauge extends Meter
{
    public static string $unit = 'g';

    public int $level = 0 {
        set {
            Gauge::$unit = 'mg';
            <weak_warning descr="Modify this static property only from the class that declares it.">parent::$limit = 5</weak_warning>;
            <weak_warning descr="Modify this static property only from the class that declares it.">static::$limit = 6</weak_warning>;
            $this->level = $value;
        }
    }
}

function probe()
{
    return new class (<weak_warning descr="Modify this static property only from the class that declares it.">Meter::$limit = 1</weak_warning>) {
        public function __construct(int $n)
        {
        }
    };
}

new class (<weak_warning descr="Modify this static property only from the class that declares it.">Meter::$limit = 2</weak_warning>) {
    public function __construct(int $n)
    {
    }
};
