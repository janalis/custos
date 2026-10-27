<?php
class Text
{
    public function indexOf(int $offset): ?int { return null; }

    public function inline(int $length): int
    {
        if (null === $length = $this->indexOf($length)) {
            return 0;
        }
        return $length;
    }

    public function after(int $length): int
    {
        $length = $this->indexOf($length);
        if ($length == null) {
            throw new \RuntimeException('none');
        }
        return $length;
    }

    public function loop(array $rows, int $n): void
    {
        foreach ($rows as $r) {
            $n = $this->indexOf($r);
            if (null === $n) continue;
            if (null === ($n = $this->indexOf($n))) break;
            if ($n === null) exit(1);
        }
    }

    public function kept(int $length): int
    {
        $length = <warning descr="Assigning a value of type null does not match the parameter's declared type.">$this->indexOf($length)</warning>;
        if (null === $length) {
            echo 'none';
        }
        if (null === $length = <warning descr="Assigning a value of type null does not match the parameter's declared type.">$this->indexOf($length)</warning>) {
        }
        if (null === $length = <warning descr="Assigning a value of type null does not match the parameter's declared type.">$this->indexOf($length)</warning>) {
            return 0;
        } else {
            echo 1;
        }
        $length = <warning descr="Assigning a value of type null does not match the parameter's declared type.">$this->indexOf($length)</warning>;
        if ($length !== null) {
            return 0;
        }
        $length = <warning descr="Assigning a value of type null does not match the parameter's declared type.">$this->indexOf($length)</warning>;
        if (0 === $length) {
            return 0;
        }
        return $length;
    }

    public function last(int $length): void
    {
        echo $length = <warning descr="Assigning a value of type null does not match the parameter's declared type.">$this->indexOf($length)</warning>;
        $length = <warning descr="Assigning a value of type null does not match the parameter's declared type.">$this->indexOf($length)</warning>;
    }
}
