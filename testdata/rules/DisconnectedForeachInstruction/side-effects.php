<?php
class Report
{
    public function build(array $rows, \Progress $meter, \SplStack $pending): void
    {
        foreach ($rows as $row) {
            $this->write($row);
            $meter->tick();
            $pending->pop();
        }
        foreach ($rows as $row) {
            $this->write($row);
            <weak_warning descr="Statement does not depend on the loop; move it out.">echo $meter->label();</weak_warning>
        }
    }

    private function write(array $row): void {}
}
