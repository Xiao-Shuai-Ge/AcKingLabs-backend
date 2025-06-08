package configs

var Conf = new(Config)

type Config struct {
	App           ApplicationConfig   `mapstructure:"app"`
	Log           LoggerConfig        `mapstructure:"log"`
	DB            DBConfig            `mapstructure:"database"`
	Redis         RedisConfig         `mapstructure:"redis"`
	Email         EmailConfig         `mapstructure:"email"`
	JWT           JWTConfig           `mapstructure:"jwt"`
	Oss           OssConfig           `mapstructure:"oss"`
	Elasticsearch ElasticsearchConfig `mapstructure:"elasticsearch"`
	Ulearning     UlearningConfig     `mapstructure:"ulearning"`
}

type ApplicationConfig struct {
	Host        string `mapstructure:"host"`
	Port        int    `mapstructure:"port"`
	Env         string `mapstructure:"env"`
	LogfilePath string `mapstructure:"logfilePath"`
}
type LoggerConfig struct {
	Level    int8   `mapstructure:"level"`
	Format   string `mapstructure:"format"`
	Director string `mapstructure:"director"`
	ShowLine bool   `mapstructure:"show-line"`
}

type DBConfig struct {
	Driver      string `mapstructure:"driver"`
	AutoMigrate bool   `mapstructure:"migrate"`
	Dsn         string `mapstructure:"dsn"`
}
type RedisConfig struct {
	Enable   bool   `mapstructure:"enable"`
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
}

type EmailConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	UserName string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type JWTConfig struct {
	Secret string `mapstructure:"secret"`
}

type KafkaConfig struct {
	host string `mapstructure:"host"`
	port int    `mapstructure:"port"`
}

type OssConfig struct {
	Endpoint        string `mapstructure:"endpoint"`
	AccessKeyID     string `mapstructure:"accessKeyID"`
	AccessKeySecret string `mapstructure:"accessKeySecret"`
	BucketName      string `mapstructure:"bucketName"`
}

type ElasticsearchConfig struct {
	Host     string `mapstructure:"host"`
	Port     string `mapstructure:"port"`
	UserName string `mapstructure:"username"`
	Password string `mapstructure:"password"`
}

type UlearningConfig struct {
	SigninOperation        string `mapstructure:"signinOperation"`
	GetSigninDetail        string `mapstructure:"getSigninDetail"`
	GetCourseActivities    string `mapstructure:"getCourseActivities"`
	GetAllCourses          string `mapstructure:"getAllCourses"`
	GetActivityDetail      string `mapstructure:"getActivityDetail"`
	SigninTeacherOperation string `mapstructure:"signinTeacherOperation"`
	Login                  string `mapstructure:"login"`
	Teacher                string `mapstructure:"teacher"`
	Password               string `mapstructure:"password"`
}
