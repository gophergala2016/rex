lark.task{'test', function()
    lark.exec{'sh', '-c', 'go test -cover $(glide novendor)'}
end}

lark.task{'push', function()
    lark.exec{'git', 'push', 'upstream', 'HEAD', ignore=true}
    lark.exec{'git', 'push', 'origin', 'HEAD', ignore=true}
end}
