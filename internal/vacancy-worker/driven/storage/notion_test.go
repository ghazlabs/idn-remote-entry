package storage_test

import (
	"context"
	"testing"

	"github.com/ghazlabs/idn-remote-entry/internal/shared/core"
	"github.com/ghazlabs/idn-remote-entry/internal/testutil"
	"github.com/ghazlabs/idn-remote-entry/internal/vacancy-worker/driven/storage"
	"github.com/go-resty/resty/v2"
	"github.com/riandyrn/go-env"
	"github.com/stretchr/testify/require"
)

func TestSaveReallyLongJobDescription(t *testing.T) {
	s, err := storage.NewNotionStorage(storage.NotionStorageConfig{
		DatabaseID:  env.GetString(testutil.EnvKeyNotionDatabaseID),
		NotionToken: env.GetString(testutil.EnvKeyNotionToken),
		HttpClient:  resty.New(),
	})
	require.NoError(t, err)

	vacancyLongDesc := core.Vacancy{
		JobTitle:        "Testing Vacancy With Really Long Job Description",
		CompanyName:     "Testing Company",
		CompanyLocation: "Testing Location",
		RelevantTags:    []string{"testing", "long", "description"},
		ApplyURL:        "https://example.com",
		ShortDescription: `Lorem ipsum dolor sit amet, consectetur adipiscing elit. Etiam congue ut tortor eget tincidunt. Duis ligula tellus, eleifend ac placerat a, consequat vitae lorem. Sed finibus accumsan orci tristique ultrices. Duis egestas lorem tellus, at tincidunt leo accumsan ac. Maecenas aliquet leo ut odio sodales maximus. Donec augue elit, volutpat ac enim sed, tincidunt malesuada neque. Praesent consequat facilisis scelerisque. Aliquam erat volutpat. Aliquam nulla nisl, lobortis vitae elit vel, semper cursus sem. Donec feugiat arcu quis consequat pellentesque. Ut sagittis tempor ipsum, in suscipit sapien imperdiet vitae. Aenean eu ornare lectus, pulvinar pellentesque erat. Suspendisse semper, purus a consectetur convallis, magna elit varius lacus, vitae consectetur tortor sapien in dui. Nunc laoreet dolor vel tellus feugiat, ac posuere orci feugiat. Integer condimentum rhoncus bibendum.

Curabitur vel mauris quis nisi tincidunt pellentesque eu nec augue. Praesent vel sapien ante. Nam quis elit sed est auctor rutrum eu et dui. Praesent ut maximus lectus. Duis sapien dolor, interdum at venenatis id, porta quis nisi. Nam lobortis eros in sem cursus, sit amet vestibulum lorem luctus. Suspendisse eleifend lectus at mauris congue blandit. Nunc orci lorem, commodo quis gravida in, lobortis egestas diam.

Praesent pharetra est leo, ut ornare justo molestie sed. Ut ut sagittis nibh, nec faucibus ipsum. Maecenas vitae tincidunt orci, nec dictum sapien. Cras ac scelerisque purus. Sed vulputate posuere finibus. Sed rutrum consectetur ipsum sed faucibus. Mauris ultrices in tellus sed pulvinar. Nunc imperdiet tellus ac metus viverra suscipit. Aliquam nibh ipsum, vulputate sit amet malesuada ut, elementum at libero. Cras sed justo sit amet enim dapibus venenatis quis non mauris. Mauris eu odio a odio lacinia finibus eu vel risus. Nulla mattis non ipsum non porttitor.

Donec consequat nibh in viverra rhoncus. Nulla facilisi. In pharetra turpis risus, eu volutpat massa scelerisque semper. Duis malesuada rutrum semper. Fusce varius mauris quis eros maximus, eget bibendum ex eleifend. Phasellus vitae nisi auctor ipsum rhoncus aliquam. Phasellus cursus eget tortor eget condimentum. Phasellus sapien purus, finibus sit amet molestie non, dictum id purus. Aliquam vel lorem sagittis dolor congue tempor.

Phasellus non mauris ullamcorper, tempor erat dictum, tempus nulla. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Pellentesque non diam porta, posuere ligula ut, lobortis lectus. Donec vel nisl urna. Maecenas lobortis odio non turpis aliquam consequat. Nam ac vulputate arcu. Nullam ac hendrerit augue. Nunc tincidunt blandit leo ac vehicula. Aenean finibus, tortor ac sollicitudin convallis, tellus augue varius nunc, ut ultrices dolor nisi in lectus. Fusce erat elit, fermentum hendrerit sem nec, imperdiet mollis orci.

Nunc tincidunt erat et consectetur dictum. Donec sagittis leo lacus, nec sollicitudin massa lobortis id. Nulla iaculis urna eget libero pulvinar posuere. Donec vestibulum et nulla sit amet efficitur. Proin tincidunt ipsum et ullamcorper pulvinar. Fusce accumsan laoreet pulvinar. Pellentesque accumsan posuere leo quis rhoncus. Ut quis leo sit amet elit finibus tempus. Maecenas ut porta nisi. Quisque varius justo ac feugiat viverra. Aenean ligula est, tristique nec leo non, porta luctus odio. Donec nec odio id lorem tempor maximus eu convallis dui. Nulla nec eros at sapien ultrices feugiat. Suspendisse congue maximus feugiat.

Donec id nibh ac erat pulvinar fermentum et id dolor. Morbi faucibus libero mauris, vestibulum pretium ligula sodales et. Praesent eu imperdiet libero. Maecenas sit amet ligula ut urna mollis imperdiet vitae ut orci. Vivamus elementum tempor risus, finibus vulputate neque feugiat et. Vivamus quis est sem. Morbi volutpat velit in risus commodo, eget laoreet felis venenatis. Aliquam neque purus, commodo vel dapibus et, dictum a ex. Sed sit amet nisl ac turpis dignissim vehicula. In vitae fermentum leo. Ut sodales, ligula eu luctus egestas, nisi tortor elementum elit, sit amet cursus libero turpis eget mauris. Cras lorem mauris, dapibus maximus tempor eu, mattis nec sem. In sollicitudin interdum lorem non vestibulum.

Sed tincidunt tempor orci. Curabitur ac molestie felis. Proin eu congue tortor. Aenean mi nibh, porttitor id leo euismod, mollis ullamcorper felis. Vivamus convallis augue id neque aliquet, nec rutrum ex volutpat. Quisque pharetra massa mauris, eu fermentum velit interdum sit amet. Curabitur aliquet non neque ut ultrices. Donec dolor velit, sagittis eget nisi sit amet, laoreet viverra tellus. Morbi fringilla lectus vel finibus tempor. Morbi porttitor venenatis leo ut interdum. Pellentesque vitae sem justo.

Fusce volutpat fermentum massa a suscipit. Nunc varius felis sagittis, gravida massa sed, ultrices nulla. Quisque non ipsum lacinia lorem elementum tristique eget ut turpis. Phasellus ornare tempus erat sit amet dapibus. Pellentesque eu elit id nibh hendrerit lobortis. Vestibulum id arcu dui. Aliquam venenatis mauris turpis, vel rutrum est cursus eget. Etiam lobortis interdum placerat. Cras varius vestibulum aliquet. Pellentesque habitant morbi tristique senectus et netus et malesuada fames ac turpis egestas. In fermentum dictum nisl, a aliquam arcu maximus nec. Vestibulum maximus quam sit amet leo consequat, id rutrum nunc feugiat.

Curabitur in turpis a orci varius elementum. Duis augue urna, vulputate ultricies sodales ut, laoreet sit amet magna. Etiam a sapien ac lacus mattis mollis. Vestibulum feugiat mattis magna id posuere. Ut neque ante, tempor sit amet nisi sagittis, scelerisque consequat est. Aliquam feugiat et sapien nec porttitor. Aenean sed ante ac augue fringilla venenatis quis ac tellus. Vestibulum sagittis neque ligula, et luctus urna consequat et. Morbi sit amet pulvinar lectus. Mauris ut molestie dui. Nam finibus nulla eu nisi pharetra, nec euismod elit malesuada. Nulla non ipsum sagittis, interdum nisi sed, consectetur elit.

Praesent ut elit vitae diam tincidunt lobortis. Vestibulum mollis nisl in odio consectetur, ut tempor lectus hendrerit. Vestibulum ante ipsum primis in faucibus orci luctus et ultrices posuere cubilia curae; Donec tincidunt ut dolor a iaculis. Cras laoreet nulla et nunc maximus, in efficitur metus maximus. Sed viverra interdum interdum. Morbi et viverra lectus.

In fringilla sapien in orci dapibus, ac volutpat magna consectetur. In in turpis elit. Cras augue metus, elementum id consectetur vitae, maximus eu lorem. Fusce sollicitudin aliquam sollicitudin. Vivamus ante nulla, porta non tortor suscipit, placerat tempus nunc. Quisque et volutpat ex. Quisque lacus odio, bibendum pulvinar mi at, faucibus pretium massa. Vestibulum id purus feugiat metus viverra finibus id non est. Integer felis felis, molestie et consectetur nec, commodo eget enim. Aliquam eu orci sem.

Donec malesuada nisl non massa tristique mattis. Curabitur ac cursus velit. Ut posuere dui a sapien molestie bibendum. Cras pulvinar elit nisl. Aliquam ac diam convallis, tincidunt ante in, mollis justo. Fusce blandit diam felis, accumsan laoreet augue pretium non. Sed blandit odio lacus, nec vehicula justo rhoncus sed. Quisque semper vulputate dolor, sit amet fermentum augue tempor eget.

Aliquam consectetur mi eu justo malesuada pharetra. Donec laoreet tincidunt mauris quis aliquet. Aenean neque velit, porttitor non tincidunt eget, viverra a ligula. Pellentesque rhoncus eget nulla in mollis. Nulla congue ante at magna tristique auctor. Cras semper nisi id odio ultricies volutpat. Aliquam consectetur maximus risus, in iaculis ante porta non. Fusce hendrerit sagittis justo at commodo. Sed et vestibulum nulla. Donec ac arcu et erat porta venenatis. Quisque imperdiet luctus elementum.

Maecenas nec dictum justo. Mauris velit nisl, semper vulputate varius quis, ornare quis leo. Praesent lacinia sem vel dui porttitor mollis vitae sed sapien. In volutpat sed sem vitae interdum. Sed ut posuere nisi, sit amet egestas mi. Morbi euismod sed libero at varius. Aenean tempor elit magna, vel mollis lectus consectetur ac. Aenean a enim ut felis hendrerit fermentum a sed eros. Sed congue, sapien ac porttitor congue, justo ipsum luctus metus, non venenatis metus justo scelerisque lacus. Proin tempor vehicula enim. Etiam eu egestas ligula. Curabitur sit amet tellus aliquam arcu consequat blandit ac a nibh.

Class aptent taciti sociosqu ad litora torquent per conubia nostra, per inceptos himenaeos. Sed consectetur nec dolor non pharetra. Etiam ullamcorper nisi eu velit facilisis, sed accumsan arcu vulputate. Ut erat dolor, laoreet at posuere quis, efficitur at lectus. Nunc id commodo metus. Sed rhoncus sapien vitae nibh hendrerit, vitae rutrum velit ornare. Maecenas ut odio ante.

Donec varius velit in erat vestibulum, eget finibus sapien ornare. Quisque sagittis orci quis blandit pulvinar. Nulla nec justo sit amet tellus efficitur gravida. Cras hendrerit in diam vitae egestas. Nullam venenatis risus nec nulla tincidunt facilisis. Integer quis porttitor sapien. Praesent eget dolor et velit elementum malesuada at sed justo. Morbi et tortor elit. Ut imperdiet vehicula blandit.

Aliquam erat volutpat. Mauris ornare scelerisque felis, in tristique mauris pulvinar et. Vestibulum leo est, malesuada nec ligula non, elementum sodales nunc. Vivamus urna elit, molestie eu magna sed, rutrum volutpat ante. Aliquam mattis vehicula sem, eu pulvinar nisi dictum quis. Duis velit tortor, varius ut bibendum non, volutpat at magna. In congue porta nulla quis fermentum.

Donec eget auctor nunc. Phasellus ullamcorper odio eu odio ultrices maximus. Vestibulum ac odio arcu. Donec pellentesque non risus eget feugiat. Lorem ipsum dolor sit amet, consectetur adipiscing elit. Vestibulum eu lacinia diam, sit amet auctor augue. In in eros in nisl vestibulum euismod sit amet sit amet nibh. Aenean sit amet sem eu sapien semper porttitor. Nam vitae ligula at erat finibus ultrices. Fusce faucibus nisl eu lorem imperdiet, non gravida blandit.`,
	}

	rec, err := s.Save(context.Background(), vacancyLongDesc)
	require.NoError(t, err)

	require.NotEmpty(t, rec.ID)
}
