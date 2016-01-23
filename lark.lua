lark.task{'test', function()
    lark.exec{'go', 'test', '-cover', './...'}
end}
