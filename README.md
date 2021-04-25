# Week 2

Q:
我们在数据库操作的时候，比如 dao 层中当遇到一个 sql.ErrNoRows 的时候，是否应该 Wrap 这个 error，抛给上层。为什么，应该怎么做请写出代码？

A:




# test
```bash
curl http://127.0.0.0:8080/product/A2

curl -X POST http://127.0.0.1:8080/product/update -d '{"code": "A2", "price": 130}'
```

