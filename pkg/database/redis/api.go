/**
 * redis 通用操作接口
 * @author yinjk
 * @create 2019-02-11 14:16
 */
package redis

type API interface {

	/**
	 * 关闭连接
	 * @param :
	 * @return:
	 * @author: yinjk
	 * @time  : 2019/2/19 17:53
	 */
	Close() error

	/**
	 * 开启redis事物
	 * @param :
	 * @return:
	 * @author: yinjk
	 * @time  : 2019/2/13 14:53
	 */
	Multi()

	/**
	 * 执行事物块内的所有语句
	 * @param :
	 * @return:
	 * @author: yinjk
	 * @time  : 2019/2/13 14:53
	 */
	Exec()

	/**
	 * 取消事物，放弃事物块内的所有语句
	 * @param :
	 * @return:
	 * @author: yinjk
	 * @time  : 2019/2/13 14:53
	 */
	Discard()

	/**
	 * 删除一个或多个key
	 * @param : 要删除的key
	 * @return: 删除成功的个数
	 * @author: yinjk
	 * @time  : 2019/2/13 10:39
	 */
	DeleteKey(keys ...interface{}) int

	/**
	 * key是否在redis中存在
	 * @param : key
	 * @return: 是否存在
	 * @author: yinjk
	 * @time  : 2019/2/13 10:40
	 */
	Exists(key string) bool

	/**
	 * 获取一个string类型的值
	 * @param : key
	 * @return:
	 * @author: yinjk
	 * @time  : 2019/2/13 10:40
	 */
	GetString(key string) string

	/**
	 * 获取一个int类型的值
	 * @param : key
	 * @return:
	 * @author: yinjk
	 * @time  : 2019/2/13 10:41
	 */
	GetInt(key string) (value int)

	/**
	 * 设置一个key-value值
	 * @param :
	 * @return: 是否设置成功
	 * @author: yinjk
	 * @time  : 2019/2/13 10:41
	 */
	Set(key, value interface{}) bool

	/**
	 * @description: 设置key过期时间
	 * @param
	 * @return
	 * @author: wzl
	 * @time  : 2019/8/23 16:12
	 */
	Expire(key string, time int) bool

	/**
	 * 向一个map中设置一个value值
	 * @param : key map的键
	 * @param : filed map中的字段
	 * @param : value map中字段的值
	 * @return: 是否设置成功
	 * @author: yinjk
	 * @time  : 2019/2/13 10:42
	 */
	HSet(key, field string, value interface{}) bool

	/**
	 * 从map中获取一个字段的值
	 * @param : key 要获取map的键
	 * @param : field 要获取map的字段
	 * @return: 获取的值
	 * @author: yinjk
	 * @time  : 2019/2/13 10:43
	 */
	HGet(key, field string) string

	/**
	 * 获取一整个map所有字段的值
	 * @param : 要获取的map
	 * @return: 整个map的值
	 * @author: yinjk
	 * @time  : 2019/2/13 10:44
	 */
	HGetAll(key string) map[string]string

	/**
	 * 设置整个map的值
	 * @param : key map的键
	 * @return: map的数据
	 * @author: yinjk
	 * @time  : 2019/2/13 10:45
	 */
	HmSet(key string, data map[string]interface{})

	/**
	 * 想有序集合中追加一个值
	 * @param : key 有序集合的键
	 * @param : score 要追加的值的排序分数
	 * @param : value 要追加的值
	 * @return:
	 * @author: yinjk
	 * @time  : 2019/2/13 10:46
	 */
	ZAdd(key string, score int, value interface{}) error

	/**
	 * 获取有序集合中的元素个数
	 * @param :
	 * @return:
	 * @author: yinjk
	 * @time  : 2019/2/13 13:48
	 */
	ZCard(key string) int

	/**
	 * 获取有序集合中某个值的排序分数
	 * @param : key 有序集合的键
	 * @param : member 要获取分数的值
	 * @return: 分数
	 * @author: yinjk
	 * @time  : 2019/2/13 10:47
	 */
	ZScore(key, member string) int

	/**
	 * 移除有序集合中的一个或多个成员
	 * @param : 集合的key
	 * @param : 集合key中的值
	 * @return: 移除的值的个数
	 * @author: yinjk
	 * @time  : 2019/2/11 15:25
	 */
	ZRemove(key string, member ...interface{}) int

	/**
	 * 向有序集合中追加数据
	 * @param : key 有序集合的键
	 * @param : values 要追加的values
	 * @return: 是否最佳成功
	 * @author: yinjk
	 * @time  : 2019/2/11 16:45
	 */
	ZAddSet(key string, values []interface{}) bool

	/**
	 * 清空有序集合并重新添加一组数据（分数按照 0 1 2 3... 排序）
	 * @param : key 有序集合的key
	 * @param : values 要添加的数据
	 * @return: 是否添加成功
	 * @author: yinjk
	 * @time  : 2019/2/11 16:42
	 */
	ZClearAndAddSet(key string, values []interface{}) bool

	/**
	 * 获取有序集合中的最大分数
	 * @param : 有序集合的key
	 * @return: 最大的分数
	 * @author: yinjk
	 * @time  : 2019/2/11 16:42
	 */
	ZGetMaxScore(key string) int

	/**
	 * 通过索引区间返回有序集合成指定区间内的成员，分数从低到高
	 * @param : key 有序集合的键
	 * @param : cronjob 从第几个元素开始（从0开始计数）
	 * @param : end 到第几个元素结束（-1 表示到最后）
	 * @return: 有序集合
	 * @author: yinjk
	 * @time  : 2019/2/11 16:31
	 */
	ZRange(key string, start, stop int) []ZSetValue

	/**
	 * 返回有序集中指定区间内的成员，通过索引，分数从高到底，相当于ZRange倒序
	 * @param :
	 * @return:
	 * @author: yinjk
	 * @time  : 2019/2/11 16:38
	 */
	ZRevRange(key string, start, stop int) []ZSetValue
}
